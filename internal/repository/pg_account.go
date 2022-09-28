package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/alexveli/astral-praktika/internal/domain"
	"github.com/alexveli/astral-praktika/internal/proto"
	mylog "github.com/alexveli/astral-praktika/pkg/log"
)

type AccountRepo struct {
	conn *pgxpool.Pool
}

func NewAccountRepo(db *pgxpool.Pool) *AccountRepo {
	return &AccountRepo{
		conn: db,
	}
}

func (a *AccountRepo) GetAccount(ctx context.Context, login *proto.Account) (*proto.Account, error) {
	var uuid sql.NullInt64
	var uname, pwd, refreshToken sql.NullString
	var expiresAt sql.NullTime
	row := a.conn.QueryRow(ctx, `SELECT uuid, username, passwordhash, refresh_token, expires_at FROM accounts 
				WHERE username = $1 OR uuid = $2 OR refresh_token = $3`, login.Username, login.UserID, login.RefreshToken)
	err := row.Scan(&uuid, &uname, &pwd, &refreshToken, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			mylog.SugarLogger.Errorf("no rows %v", err)

			return &proto.Account{}, domain.ErrUserNotFound
		}
		mylog.SugarLogger.Errorf("error when scanning values %v", err)

		return &proto.Account{}, err
	}
	if !(uuid.Valid && uname.Valid && pwd.Valid && refreshToken.Valid && expiresAt.Valid) {
		mylog.SugarLogger.Warnf("account fields are invalid %v", err)

		return &proto.Account{}, domain.ErrAccountFieldsInValid
	}
	return &proto.Account{
		UserID:       uuid.Int64,
		Username:     uname.String,
		PasswordHash: pwd.String,
		RefreshToken: refreshToken.String,
		ExpiresAt:    timestamppb.New(expiresAt.Time.Local()),
	}, nil
}

func (a *AccountRepo) StoreAccount(ctx context.Context, account *proto.Account) error {
	var uuid sql.NullInt64
	insertAccount := "INSERT INTO accounts (username, passwordhash) VALUES($1,$2) RETURNING uuid"
	executionResult := a.conn.QueryRow(ctx, insertAccount, account.Username, account.PasswordHash)
	err := executionResult.Scan(&uuid)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot store account: %v", err)

		return err
	}
	if uuid.Valid {
		account.UserID = uuid.Int64
	}
	mylog.SugarLogger.Infof("user %s successfully registered with id %d", account.Username, account.UserID)

	return nil
}

func (a *AccountRepo) UpdateRefreshToken(ctx context.Context, account *proto.Account) error {
	updateAccounts := "UPDATE accounts SET refresh_token = $1, expires_at = $2 WHERE uuid = $3"
	_, err := a.conn.Exec(ctx, updateAccounts, account.RefreshToken, account.ExpiresAt.AsTime(), account.UserID)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot update refresh token, %v", err)

		return err
	}
	return nil
}
