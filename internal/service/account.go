package service

import (
	"context"
	"encoding/json"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/alexveli/astral-praktika/internal/domain"
	"github.com/alexveli/astral-praktika/internal/proto"
	"github.com/alexveli/astral-praktika/internal/repository"
	"github.com/alexveli/astral-praktika/pkg/auth"
	mylog "github.com/alexveli/astral-praktika/pkg/log"
)

type AccountService struct {
	repo         repository.Authenticator
	tokenManager *auth.Manager
}

func NewAccountService(repo repository.Authenticator, tokenManager *auth.Manager) *AccountService {
	return &AccountService{repo: repo, tokenManager: tokenManager}
}

func (a *AccountService) Register(ctx context.Context, account *proto.Account) error {
	err := a.repo.StoreAccount(ctx, account)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot store user account: %v", err)
		return err
	}
	return nil
}

func (a *AccountService) Login(ctx context.Context, login *proto.Account) error {
	account, err := a.repo.GetAccount(ctx, login)
	if err != nil {

		return domain.ErrUserNotFound
	}
	err = verifyPassword(login.PasswordHash, account.PasswordHash)
	if err != nil {

		return domain.ErrPasswordIncorrect
	}
	return nil
}

func (a *AccountService) GenerateTokens(ctx context.Context, uuid int64) ([]byte, string, error) {
	accessToken, err := a.tokenManager.GenerateToken(uuid, domain.ACCESS)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot generate accessToken for user %d, %v", uuid, err)

		return []byte{}, "", err
	}
	refreshToken, err := a.tokenManager.GenerateToken(uuid, domain.REFRESH)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot generate refreshToken for user %d, %v", uuid, err)

		return []byte{}, "", err
	}
	err = a.repo.UpdateRefreshToken(ctx, &proto.Account{
		UserID:       uuid,
		RefreshToken: refreshToken.Token,
		ExpiresAt:    timestamppb.New(refreshToken.ExpiresAt.AsTime()),
	})
	if err != nil {
		mylog.SugarLogger.Errorf("cannot save refresh token, %v", err)

		return []byte{}, "", err
	}
	tokens, err := json.Marshal(proto.Tokens{
		AccessToken:  accessToken.Token,
		RefreshToken: refreshToken.Token,
	})
	if err != nil {
		mylog.SugarLogger.Errorf("cannot marshal tokens, %v", err)

		return []byte{}, "", err
	}
	return tokens, accessToken.Token, nil
}

func (a *AccountService) GetAccountFromToken(ctx context.Context, token string) (*proto.Account, error) {
	userid, err := a.tokenManager.ExtractUserIDFromToken(token)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get userid, %v", err)

		return &proto.Account{}, err
	}
	account, err := a.GetAccount(ctx, userid)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get account, %v", err)

		return &proto.Account{}, err
	}

	return account, nil
}

func (a *AccountService) RefreshToken(ctx context.Context, token string) ([]byte, string, error) {
	lookup := proto.Account{
		RefreshToken: token,
	}
	account, err := a.repo.GetAccount(ctx, &lookup)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get account, %v", err)

		return []byte{}, "", err
	}
	if tokenExpired(account.ExpiresAt) {

		return []byte{}, "", domain.ErrAccountExpired
	}

	return a.GenerateTokens(ctx, account.UserID)
}

func (a *AccountService) TokenValid(token string) error {
	return a.tokenManager.TokenValid(token)
}

func (a *AccountService) GetAccount(ctx context.Context, userid int64) (*proto.Account, error) {
	lookup := proto.Account{
		UserID: userid,
	}
	account, err := a.repo.GetAccount(ctx, &lookup)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get account, %v", err)

		return &proto.Account{}, err
	}
	return account, nil
}

func tokenExpired(expiresAt *timestamppb.Timestamp) bool {
	return time.Now().After(expiresAt.AsTime())
}

func verifyPassword(pwd1 string, pwd2 string) error {
	if pwd1 != pwd2 {

		return domain.ErrPasswordIncorrect
	}

	return nil
}
