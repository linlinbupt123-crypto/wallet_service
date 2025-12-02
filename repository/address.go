package repository

import (
	"context"

	"github.com/linlinbupt123-crypto/wallet_service/entity"
)

type AddressRepo interface {
	SaveAddress(ctx context.Context, addr *entity.Address) error
	GetAddressesByUser(ctx context.Context, userID string) ([]*entity.Address, error)
}
