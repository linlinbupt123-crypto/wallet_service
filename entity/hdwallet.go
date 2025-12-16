package entity

import (
	"time"
)

type HDWallet struct {
	ID                string    `bson:"id"`
	UserID            string    `bson:"user_id"`
	MnemonicEncrypted []byte    `bson:"mnemonic_encrypted"`
	EncryptedSeed     []byte    `bson:"encrypted_seed"`
	XPrvEncrypted     []byte    `bson:"xprv_encrypted"`
	XPub              string    `bson:"xpub"`
	SaltHex           string    `bson:"salt_hex"`
	CreatedAt         time.Time `bson:"created_at"`
}
