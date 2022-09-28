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

type SecretKeeperRepo struct {
	conn *pgxpool.Pool
}

func NewSecretKeeperRepo(db *pgxpool.Pool) *SecretKeeperRepo {
	return &SecretKeeperRepo{conn: db}
}

func (r *SecretKeeperRepo) IncrementUsage(ctx context.Context, secret *proto.Secret) error {
	insertAccesses := "INSERT INTO accesses (uuid, secretid, accessed_at) VALUES($1, $2, $3) RETURNING accessid"
	_, err := r.conn.Exec(
		ctx,
		insertAccesses,
		secret.Uuid,
		secret.Secretid,
		timestamppb.Now().AsTime(),
	)
	if err != nil {
		mylog.SugarLogger.Errorf("error inserting access record for secret %v, %v", secret, err)

		return err
	}

	return nil
}

func (r *SecretKeeperRepo) GetAccessCount(ctx context.Context, secret *proto.Secret) (int64, error) {
	var selectAccesses string
	var count sql.NullInt64
	selectAccesses = "SELECT COUNT(*) FROM accesses WHERE secretid = $1"
	row := r.conn.QueryRow(ctx, selectAccesses, secret.Secretid)
	err := row.Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			mylog.SugarLogger.Infof("secretid never used,%d", secret.Secretid)

			return 0, nil
		}
		mylog.SugarLogger.Errorf("error getting count of accesses for secretid %d, %v", secret.Secretid, err)

		return 0, err
	}
	if !count.Valid {
		mylog.SugarLogger.Errorf("error getting count of accesses for secretid %d, %v", secret.Secretid, err)

		return 0, domain.ErrSecretAccessesCountInValid
	}
	return count.Int64, nil
}

func (r *SecretKeeperRepo) GetSecretByKey(ctx context.Context, key string) (*proto.Secret, error) {
	var secret sql.NullString
	var createdAt sql.NullTime
	var secretid, uuid sql.NullInt64
	selectSecret := "SELECT secretid, uuid, secret, created_at FROM secrets WHERE expired = false AND key = $1"
	row := r.conn.QueryRow(ctx, selectSecret, key)
	err := row.Scan(&secretid, &uuid, &secret, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			mylog.SugarLogger.Infof("no secret for key %s found", key)

			return &proto.Secret{}, domain.ErrSecretNoSecretForUser
		}
		mylog.SugarLogger.Errorf("error scanning secret, %v", err)

		return &proto.Secret{}, err
	}

	if !(secretid.Valid && uuid.Valid &&
		secret.Valid && createdAt.Valid) {
		mylog.SugarLogger.Infof("Error %s for key %s", domain.ErrSecretFieldsAreNotValid.Error(), key)

		return &proto.Secret{}, domain.ErrSecretFieldsAreNotValid
	}

	return &proto.Secret{
		Secretid:  secretid.Int64,
		Uuid:      uuid.Int64,
		Secret:    secret.String,
		CreatedAt: timestamppb.New(createdAt.Time.Local()),
	}, nil
}

func (r *SecretKeeperRepo) StoreSecret(ctx context.Context, secret *proto.Secret, expireExisting bool) (int64, error) {
	var secretid sql.NullInt64
	var insertSecret, expireExistingSecrets string
	tx, err := r.conn.BeginTx(ctx, pgx.TxOptions{})
	defer func() {
		if err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				mylog.SugarLogger.Errorf("cannot rollback transaction, %v", err)

				return
			}
		} else {
			err := tx.Commit(ctx)
			if err != nil {
				mylog.SugarLogger.Errorf("cannot commit transaction, %v", err)

				return
			}
		}
	}()
	insertSecret = "INSERT INTO secrets (uuid, key, secret, expired, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING secretid"
	row := tx.QueryRow(ctx, insertSecret, secret.Uuid, secret.Key, secret.Secret, secret.Expired, secret.CreatedAt.AsTime())
	err = row.Scan(&secretid)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot scan uuid, %v", err)

		return 0, err
	}
	if !secretid.Valid {
		mylog.SugarLogger.Warnf("returned secretid is null or invalid, %v", secret)

		return 0, domain.ErrSecretNotValid
	}
	if expireExisting {
		expireExistingSecrets = "UPDATE secrets SET expired=true WHERE expired=false AND uuid=$1 AND secretid<>$2"
		_, err := tx.Exec(ctx, expireExistingSecrets, secret.Uuid, secretid.Int64)
		if err != nil {
			mylog.SugarLogger.Errorf("cannot expire secrets for uuid %d, %v", secret.Uuid, err)

			return 0, err
		}
	}
	mylog.SugarLogger.Infof("secret successfully saved, %v", secret)

	return secretid.Int64, nil
}

func (r *SecretKeeperRepo) RetireSecretByID(ctx context.Context, secretid int64) bool {
	updateSecret := "UPDATE secrets SET expired = true WHERE secretid = $1"
	_, err := r.conn.Exec(ctx, updateSecret, secretid)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot expire secret, %v", err)

		return false
	}
	return true
}
