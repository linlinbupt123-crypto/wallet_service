package domain

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"

	bip39 "github.com/tyler-smith/go-bip39"

	"github.com/ethereum/go-ethereum/crypto"

	"golang.org/x/crypto/pbkdf2"

	"github.com/linlinbupt123-crypto/wallet_service/entity"
	"github.com/linlinbupt123-crypto/wallet_service/repository"
)

// NOTE:
// - This file standardizes on btcsuite's hdkeychain for BIP32/BIP44 derivation.
// - KDF metadata is encoded into SaltHex as: "pbkdf2$<iterations>$<hexsalt>"
//   so we don't need to modify entity.HDWallet struct to store algorithm/params.
// - We use AES-GCM for authenticated encryption, and PBKDF2 (sha256) with 310_000 iterations.
// - BIP39 passphrase (the optional additional mnemonic passphrase) is NOT stored here.
//   If you want to support BIP39 passphrase, accept it as a parameter and record a flag.

const (
	// KDF algorithm label used in SaltHex metadata
	kdfLabel      = "pbkdf2"
	kdfIterations = 310_000 // recommended minimum; tune for your environment
)

// ---------- Helpers ----------
func clearBytes(b []byte) {
	if b == nil {
		return
	}
	for i := range b {
		b[i] = 0
	}
}

// deriveKey derives a 32-byte AES key from passphrase+salt using PBKDF2-SHA256.
// We return a copy which the caller must clear after use.
func deriveKey(passphrase string, salt []byte, iterations int) []byte {
	return pbkdf2.Key([]byte(passphrase), salt, iterations, 32, sha256.New)
}

// encrypt uses AES-256-GCM and returns nonce|ciphertext
func encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, data, nil)
	out := append(nonce, ciphertext...)
	return out, nil
}

// decrypt expects input nonce|ciphertext
func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce := ciphertext[:nonceSize]
	ct := ciphertext[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		// do not return the raw crypto error to caller in production; wrap it.
		return nil, errors.New("failed to decrypt data")
	}
	return plain, nil
}

// encodeSaltMeta packs algorithm, iterations and salt into a single string stored in DB
// Format: "<kdfLabel>$<iterations>$<hex-salt>"
func encodeSaltMeta(salt []byte, iterations int) string {
	return fmt.Sprintf("%s$%d$%s", kdfLabel, iterations, hex.EncodeToString(salt))
}

// decodeSaltMeta parses the stored SaltHex format and returns (salt, iterations, error)
func decodeSaltMeta(meta string) ([]byte, int, error) {
	parts := strings.Split(meta, "$")
	if len(parts) != 3 {
		return nil, 0, errors.New("invalid salt metadata format")
	}
	if parts[0] != kdfLabel {
		return nil, 0, errors.New("unsupported kdf")
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, 0, errors.New("invalid kdf iterations")
	}
	salt, err := hex.DecodeString(parts[2])
	if err != nil {
		return nil, 0, errors.New("invalid salt hex")
	}
	return salt, iter, nil
}

// ---------- Wallet service ----------
type HDWallet struct {
	WalletRepo *repository.Wallet
}

func NewHDWallet() *HDWallet {
	return &HDWallet{WalletRepo: repository.NewWalletRepo()}
}

/*
CreateWallet generates mnemonic, seed, master xprv/xpub, encrypts and persists.

Parameters:
  - userID: application user id to associate wallet with
  - passphrase: the user's password used to derive the encryption key (NOT BIP39 passphrase)

Returns:
  - entity.HDWallet (persisted)

Notes:
  - We encode KDF algorithm and iterations into SaltHex so callers can later derive correctly.
  - We try to zero sensitive variables as soon as possible.
*/
func (s *HDWallet) CreateWallet(ctx context.Context, userID string, passphrase string) (*entity.HDWallet, error) {
	// 1) generate mnemonic (entropy 256 bits => 24 words)
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate entropy: %w", err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	// NOTE: we treat BIP39 passphrase (optional extra mnemonic passphrase) as out-of-band.
	// Here we use empty BIP39 passphrase. If you want to support it, accept it as param.
	seed := bip39.NewSeed(mnemonic, "")

	// 2) create master key (xprv/xpub) using btcsuite hdkeychain
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		// zero seed before returning
		clearBytes(seed)
		return nil, fmt.Errorf("failed to create master key: %w", err)
	}

	// xprv string and xpub string
	xprvStr := masterKey.String()
	xpubKey, err := masterKey.Neuter()
	if err != nil {
		clearBytes(seed)
		return nil, fmt.Errorf("failed to neuter master key: %w", err)
	}
	xpubStr := xpubKey.String()

	// 3) prepare salt + kdf params
	salt := make([]byte, 16)
	_, err = rand.Read(salt)
	if err != nil {
		clearBytes(seed)
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	// encode metadata into SaltHex for future-proofing
	saltMeta := encodeSaltMeta(salt, kdfIterations)

	// 4) derive AES key
	key := deriveKey(passphrase, salt, kdfIterations)
	// ensure we clear derived key when done
	defer clearBytes(key)

	// 5) encrypt seed, xprv, mnemonic
	encSeed, err := encrypt(seed, key)
	// seed MUST be cleared ASAP
	clearBytes(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt seed: %w", err)
	}

	encXPrv, err := encrypt([]byte(xprvStr), key)
	// attempt to clear xprv bytes - xprvStr is a string (immutable), best effort:
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt xprv: %w", err)
	}

	encMnemonic, err := encrypt([]byte(mnemonic), key)
	// clear mnemonic string bytes: convert to []byte copy then zero
	clearBytes([]byte(mnemonic)) // best-effort (does not zero the original string memory)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt mnemonic: %w", err)
	}

	// 6) assemble entity and persist
	wallet := &entity.HDWallet{
		UserID:            userID,
		MnemonicEncrypted: encMnemonic,
		EncryptedSeed:     encSeed,
		XPrvEncrypted:     encXPrv,
		XPub:              xpubStr,
		SaltHex:           saltMeta, // contains KDF metadata + hex salt
		CreatedAt:         time.Now(),
	}

	_, err = s.WalletRepo.Create(ctx, wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to persist wallet: %w", err)
	}

	return wallet, nil
}

// DecryptSeed decrypts the stored seed using the provided passphrase.
// Returns plain seed bytes which caller should clear as soon as possible.
func (s *HDWallet) DecryptSeed(wallet *entity.HDWallet, passphrase string) ([]byte, error) {
	if wallet == nil {
		return nil, errors.New("wallet is nil")
	}
	// parse salt meta
	salt, iterations, err := decodeSaltMeta(wallet.SaltHex)
	if err != nil {
		return nil, fmt.Errorf("invalid salt metadata: %w", err)
	}
	key := deriveKey(passphrase, salt, iterations)
	defer clearBytes(key)

	seed, err := decrypt(wallet.EncryptedSeed, key)
	if err != nil {
		// wrap and hide crypt details
		return nil, errors.New("incorrect passphrase or corrupted data")
	}
	return seed, nil
}

// LoadWallet returns decrypted seed and decrypted xprv (both must be cleared by caller)
func (s *HDWallet) LoadWallet(ctx context.Context, userID string, passphrase string) ([]byte, []byte, error) {
	wallet, err := s.WalletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("wallet not found: %w", err)
	}
	salt, iterations, err := decodeSaltMeta(wallet.SaltHex)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid salt metadata: %w", err)
	}
	key := deriveKey(passphrase, salt, iterations)
	defer clearBytes(key)

	seed, err := decrypt(wallet.EncryptedSeed, key)
	if err != nil {
		return nil, nil, errors.New("incorrect passphrase or corrupted data")
	}

	xprv, err := decrypt(wallet.XPrvEncrypted, key)
	if err != nil {
		// clear seed before returning
		clearBytes(seed)
		return nil, nil, errors.New("incorrect passphrase or corrupted data")
	}

	return seed, xprv, nil
}

// DeriveETHAddress derives an Ethereum address from seed using a BIP32/BIP44 derivation path.
// path: "m/44'/60'/0'/0/0" or "44'/60'/0'/0/0"
func (s *HDWallet) DeriveETHKeyPair(seed []byte, path string) (*ecdsa.PrivateKey, string, error) {
	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create master key: %w", err)
	}

	indices, err := parseDerivationPath(path)
	if err != nil {
		return nil, "", fmt.Errorf("invalid derivation path: %w", err)
	}

	key := master
	for _, idx := range indices {
		key, err = key.Derive(idx)
		if err != nil {
			return nil, "", fmt.Errorf("failed to derive child key: %w", err)
		}
	}

	// get private key bytes
	priv, err := key.ECPrivKey()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get EC private key: %w", err)
	}
	// serialize private key -> 32 bytes
	privBytes := priv.Serialize()
	// convert to go-ethereum ECDSA
	ecdsaKey, err := crypto.ToECDSA(privBytes)
	// clear privBytes as soon as possible
	clearBytes(privBytes)
	if err != nil {
		return nil, "", fmt.Errorf("failed to convert to ecdsa: %w", err)
	}
	addr := crypto.PubkeyToAddress(ecdsaKey.PublicKey)
	// zero ecdsa private key D if possible (best effort)
	// Note: crypto.ToECDSA returns a pointer to ecdsa.PrivateKey whose D is a big.Int;
	// zeroing Big.Int is not trivial / fully reliable in Go, but we can try:
	// if ecdsaKey.D != nil {
	// best-effort: overwrite bytes backing D
	// (omitted here - hard in pure Go without using math/big internals)
	// }
	return ecdsaKey, addr.Hex(), nil
}

// DeriveBTCAddress derives a BTC address from seed using provided path.
// It returns a legacy/P2PKH address for m/44' paths. For bech32 (84') and p2sh-p2wpkh (49') it currently returns an error
// TODO: implement P2WPKH and P2SH-P2WPKH encoding using btcutil/txscript when needed.
func (s *HDWallet) DeriveBTCAddress(seed []byte, path string) (string, error) {
	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return "", fmt.Errorf("failed to create master key: %w", err)
	}
	indices, err := parseDerivationPath(path)
	if err != nil {
		return "", fmt.Errorf("invalid derivation path: %w", err)
	}

	// simple guard to inform about address type based on purpose field
	if strings.HasPrefix(strings.TrimSpace(path), "m/84'") || strings.HasPrefix(strings.TrimSpace(path), "M/84'") {
		return "", errors.New("bech32 (P2WPKH) address generation not implemented in this function yet")
	}
	if strings.HasPrefix(strings.TrimSpace(path), "m/49'") || strings.HasPrefix(strings.TrimSpace(path), "M/49'") {
		return "", errors.New("P2SH-P2WPKH address generation not implemented in this function yet")
	}

	key := master
	for _, idx := range indices {
		key, err = key.Derive(idx)
		if err != nil {
			return "", fmt.Errorf("failed to derive child key: %w", err)
		}
	}

	addr, err := key.Address(&chaincfg.MainNetParams)
	if err != nil {
		return "", fmt.Errorf("failed to get address: %w", err)
	}
	return addr.EncodeAddress(), nil
}

// parseDerivationPath accepts "m/44'/60'/0'/0/0" or "44'/60'/0'/0/0"
func parseDerivationPath(path string) ([]uint32, error) {
	p := strings.TrimSpace(path)
	if strings.HasPrefix(p, "m/") || strings.HasPrefix(p, "M/") {
		p = p[2:]
	}
	if p == "" {
		return nil, errors.New("empty derivation path")
	}
	parts := strings.Split(p, "/")
	indices := make([]uint32, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return nil, errors.New("invalid path segment")
		}
		hardened := strings.HasSuffix(part, "'")
		if hardened {
			part = strings.TrimSuffix(part, "'")
		}
		v, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, errors.New("invalid derivation index")
		}
		idx := uint32(v)
		if hardened {
			idx += hdkeychain.HardenedKeyStart
		}
		indices = append(indices, idx)
	}
	return indices, nil
}

// VerifyPassphrase checks whether passphrase can decrypt the stored seed.
// Returns (true, nil) if correct; (false, nil) if passphrase wrong; (false, err) for other errors.
func (s *HDWallet) VerifyPassphrase(ctx context.Context, userID string, passphrase string) (bool, error) {
	wallet, err := s.WalletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("wallet not found: %w", err)
	}
	// parse salt meta
	salt, iterations, err := decodeSaltMeta(wallet.SaltHex)
	if err != nil {
		return false, fmt.Errorf("invalid salt metadata: %w", err)
	}
	key := deriveKey(passphrase, salt, iterations)
	defer clearBytes(key)

	_, err = decrypt(wallet.EncryptedSeed, key)
	if err != nil {
		// decryption error -> either wrong passphrase or corrupted data
		// Do not leak crypto internals to caller: return false,nil for wrong passphrase
		return false, nil
	}
	return true, nil
}
