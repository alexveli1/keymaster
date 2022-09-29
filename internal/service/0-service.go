package service

import (
	"context"

	"github.com/alexveli/astral-praktika/internal/config"
	"github.com/alexveli/astral-praktika/internal/repository"
	"github.com/alexveli/astral-praktika/pkg/auth"

	"github.com/alexveli/astral-praktika/internal/proto"
)

type Authenticator interface {
	Register(ctx context.Context, input *proto.Account) error
	RefreshToken(ctx context.Context, token string) ([]byte, string, error)
	GenerateTokens(ctx context.Context, uuid int64) ([]byte, string, error)
	GetAccountFromToken(ctx context.Context, token string) (*proto.Account, error)
	GetAccount(ctx context.Context, userid int64) (*proto.Account, error)
	Login(ctx context.Context, login *proto.Account) (*proto.Account, error)
	TokenValid(token string) error
}

type SecretKeeper interface {
	GenerateSecret(ctx context.Context, account *proto.Account) ([]byte, bool)
	ProvideSecret(ctx context.Context, key string) ([]byte, int64, error)
	ValidateKey(key string) (validKey string, valid bool)
}

type Services struct {
	Authenticator Authenticator
	SecretKeeper  SecretKeeper
}

func NewServices(repositories *repository.Repositories, cfgKeeper config.KeeperConfig, tokenManager *auth.Manager) *Services {
	return &Services{
		Authenticator: NewAccountService(repositories.Authenticator, tokenManager),
		SecretKeeper:  NewSecretKeeperService(repositories.SecretKeeper, cfgKeeper),
	}
}
