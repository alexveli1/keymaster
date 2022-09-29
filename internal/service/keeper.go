package service

import (
	"context"
	"encoding/json"
	"regexp"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/alexveli/astral-praktika/internal/config"
	"github.com/alexveli/astral-praktika/internal/domain"
	"github.com/alexveli/astral-praktika/internal/proto"
	"github.com/alexveli/astral-praktika/internal/repository"
	"github.com/alexveli/astral-praktika/pkg/generator"
	mylog "github.com/alexveli/astral-praktika/pkg/log"
)

type SecretKeeperService struct {
	repo           repository.SecretKeeper
	limit          int64
	expire         time.Duration
	secretLength   int64
	expireExisting bool
	keyLength      int64
}

func NewSecretKeeperService(repo repository.SecretKeeper, cfgKeeper config.KeeperConfig) *SecretKeeperService {
	return &SecretKeeperService{
		repo:           repo,
		limit:          cfgKeeper.AccessCount,
		expire:         cfgKeeper.ExpirationPeriod,
		secretLength:   cfgKeeper.SecretLength,
		expireExisting: cfgKeeper.ExpireExisting,
		keyLength:      cfgKeeper.KeyLength,
	}
}

func (a *SecretKeeperService) ProvideSecret(ctx context.Context, key string) ([]byte, int64, error) {
	secretStruct, err := a.repo.GetSecretByKey(ctx, key)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get secret for key %s, %v", key, err)

		return nil, 0, err
	}
	if secretExpired(secretStruct.CreatedAt.AsTime(), a.expire) {
		a.repo.RetireSecretByID(ctx, secretStruct.Secretid)
		mylog.SugarLogger.Warnf("cannot provide secret key %s, since secret has expired", key)

		return nil, 0, domain.ErrSecretHasExpired
	}
	count, err := a.repo.GetAccessCount(ctx, secretStruct)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot provide secret for key %s, since cannot get access count", key)

		return nil, 0, err
	}
	if accessCountExceeded(count, a.limit) {
		a.repo.RetireSecretByID(ctx, secretStruct.Secretid)
		mylog.SugarLogger.Warnf("cannot provide secret for key %s, since access limit exceeded", key)

		return nil, 0, domain.ErrSecretAccessesCountExceeded
	}
	err = a.repo.IncrementUsage(ctx, secretStruct)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot provide secret for key %s, since cannot document access", key)

		return nil, 0, err
	}
	mylog.SugarLogger.Infof("key and secret are valid, providing")

	secret, err := json.Marshal(proto.Secret{Secret: secretStruct.Secret})
	if err != nil {
		mylog.SugarLogger.Errorf("cannot marshal secret, %v", err)

		return nil, 0, err
	}
	return secret, secretStruct.Uuid, nil
}

func (a *SecretKeeperService) GenerateSecret(ctx context.Context, account *proto.Account) ([]byte, bool) {
	secret := generator.GenerateSecret(a.secretLength)
	if !stringGenerated(secret) {
		mylog.SugarLogger.Errorf("secret generation failed for %v", account)

		return []byte{}, false
	}
	newKey := generator.GenerateKey(a.keyLength)
	if !stringGenerated(newKey) {
		mylog.SugarLogger.Warnf("key generation failed for %v", account)

		return []byte{}, false
	}
	newSecret := &proto.Secret{
		Uuid:      account.UserID,
		Key:       newKey,
		Secret:    secret,
		Expired:   false,
		CreatedAt: timestamppb.Now(),
	}
	_, err := a.repo.StoreSecret(ctx, newSecret, a.expireExisting)
	if err != nil {
		mylog.SugarLogger.Errorf("secret generation failed for account %v, %v", account, err)

		return []byte{}, false
	}
	key, err := json.Marshal(proto.Secret{Key: newKey})
	if err != nil {
		mylog.SugarLogger.Errorf("cannot marshal new key, %v", err)

		return []byte{}, false
	}
	return key, true
}

func (a *SecretKeeperService) ValidateKey(key string) (validKey string, isValid bool) {
	if len(key) != int(a.keyLength) {
		mylog.SugarLogger.Warnf("key is of incorrect length, %s - %d", key, len(key))

		return "", false
	}
	reg, err := regexp.Compile("^[A-Za-z\\d_-]*$")
	if err != nil {
		mylog.SugarLogger.Errorf("cannot compile regexp, %v", err)

		return "", false
	}
	if !reg.MatchString(key) {
		mylog.SugarLogger.Warnf("key %s doesn't match regexp", key)

		return "", false
	}

	return key, true
}

func accessCountExceeded(count int64, limit int64) bool {
	return count >= limit
}

func secretExpired(createdAt time.Time, expire time.Duration) bool {
	return time.Now().Sub(createdAt) > expire
}

func stringGenerated(str string) bool {
	return str != ""
}
