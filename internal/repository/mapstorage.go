package repository

import "github.com/alexveli/astral-praktika/pkg/storage/maps"

type MapAccountRepo struct {
	maps *maps.Maps
}

func NewMapAccountRepo(maps *maps.Maps) *MapAccountRepo {
	return &MapAccountRepo{maps: maps}
}

type MapKeeperRepo struct {
	maps *maps.Maps
}

func NewMapKeeperRepo(maps *maps.Maps) *MapKeeperRepo {
	return &MapKeeperRepo{maps: maps}
}

type MapRepositories struct {
	MapAccountRepo *MapAccountRepo
	MapKeeperRepo  *MapKeeperRepo
}

func NewMapRepositories(maps *maps.Maps) *Repositories {
	return &Repositories{
		Authenticator: NewMapAccountRepo(maps),
		SecretKeeper:  NewMapKeeperRepo(maps),
	}
}
