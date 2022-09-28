package repository

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DBManager interface {
	CreateTables(ctx context.Context) error
}

type DBCreator struct {
	db *pgxpool.Pool
}

func NewDBCreator(db *pgxpool.Pool) *DBCreator {
	return &DBCreator{db: db}
}

const creatorSQL = `create table if not exists secrets
(
    uuid       bigint    not null,
    secret     text      not null,
    secretid   serial
        constraint secrets_pk
            primary key,
    created_at timestamp not null,
    key        text      not null,
    expired    boolean   not null
);

create table if not exists accounts
(
    uuid          serial
        constraint accounts_pk
            primary key,
    username      text,
    passwordhash  text      not null,
    refresh_token text      not null,
    expires_at    timestamp not null
);

create unique index if not exists accounts_username_uindex
    on accounts (username);

create unique index if not exists accounts_uuid_uindex
    on accounts (uuid);

create table if not exists accesses
(
    uuid        integer   not null,
    secretid    integer   not null,
    accessid    serial
        constraint accesses_pk
            primary key,
    accessed_at timestamp not null
);

create unique index if not exists accesses_accessid_uindex
    on accesses (accessid);`

func (d *DBCreator) CreateTables(ctx context.Context) error {
	_, err := d.db.Exec(ctx, creatorSQL)
	if err != nil {

		return err
	}

	return nil
}
