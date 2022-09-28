package repository

import (
	"context"

	"github.com/alexveli/astral-praktika/internal/proto"
)

func (ma *MapAccountRepo) GetAccount(ctx context.Context, login *proto.Account) (*proto.Account, error) {
	return nil, nil
}
func (ma *MapAccountRepo) StoreAccount(ctx context.Context, account *proto.Account) error {
	return nil
}
func (ma *MapAccountRepo) UpdateRefreshToken(ctx context.Context, account *proto.Account) error {
	return nil
}
