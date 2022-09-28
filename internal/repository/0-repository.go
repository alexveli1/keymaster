package repository

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/alexveli/astral-praktika/internal/proto"
)

type Authenticator interface {
	GetAccount(ctx context.Context, login *proto.Account) (*proto.Account, error)
	StoreAccount(ctx context.Context, account *proto.Account) error
	UpdateRefreshToken(ctx context.Context, account *proto.Account) error
}

type SecretKeeper interface {
	IncrementUsage(ctx context.Context, secret *proto.Secret) error
	GetSecretByKey(ctx context.Context, key string) (*proto.Secret, error)
	GetAccessCount(ctx context.Context, secret *proto.Secret) (int64, error)
	StoreSecret(ctx context.Context, secret *proto.Secret, expireExisting bool) (int64, error)
	RetireSecretByID(ctx context.Context, secretid int64) bool
}

type Repositories struct {
	Authenticator Authenticator
	SecretKeeper  SecretKeeper
}

func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		Authenticator: NewAccountRepo(db),
		SecretKeeper:  NewSecretKeeperRepo(db),
	}
}
