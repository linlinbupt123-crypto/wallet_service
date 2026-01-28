package entity

import (
	"time"
)

type Wallet struct {
	ID         string `bson:"_id,omitempty"`
	UserID     string `bson:"user_id"`
	WalletName string `bson:"wallet_name"`
	WalletType string `bson:"wallet_type"` // "hd" / "imported"

	// HD 类型相关
	MnemonicEncrypted []byte `bson:"mnemonic_encrypted"`
	EncryptedSeed     []byte `bson:"encrypted_seed"`
	XPrvEncrypted     []byte `bson:"xprv_encrypted"`
	XPub              string `bson:"xpub"`

	// common 字段
	SaltHex string `bson:"salt_hex"`

	// Imported 类型相关
	CipherKey []byte `bson:"cipher_key,omitempty"` // 加密私钥

	CreatedAt time.Time `bson:"created_at"`
}
