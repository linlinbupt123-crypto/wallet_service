package entity

import (
	"time"
)

type HDWallet struct {
    UserID        string    `bson:"user_id"`
    EncryptedSeed []byte    `bson:"encrypted_seed"`
    XPrvEncrypted []byte    `bson:"xprv_encrypted"`
    XPub          string    `bson:"xpub"`
    SaltHex       string    `bson:"salt_hex"`
    CreatedAt     time.Time `bson:"created_at"`
}
