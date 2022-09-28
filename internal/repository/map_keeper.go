package repository

import (
	"context"

	"github.com/alexveli/astral-praktika/internal/proto"
)

func (mk *MapKeeperRepo) IncrementUsage(ctx context.Context, secret *proto.Secret) error {
	return nil
}
func (mk *MapKeeperRepo) GetSecretByKey(ctx context.Context, key string) (*proto.Secret, error) {
	return nil, nil
}
func (mk *MapKeeperRepo) GetAccessCount(ctx context.Context, secret *proto.Secret) (int64, error) {
	return 0, nil
}
func (mk *MapKeeperRepo) StoreSecret(ctx context.Context, secret *proto.Secret, expireExisting bool) (int64, error) {
	return 0, nil
}
func (mk *MapKeeperRepo) RetireSecretByID(ctx context.Context, secretid int64) bool {
	return false
}
