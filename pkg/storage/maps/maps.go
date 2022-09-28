package maps

import (
	"sync"

	"github.com/alexveli/astral-praktika/internal/proto"
)

type Maps struct {
	secretMap   map[int64]proto.Secret
	accessesMap map[int64]proto.Access
	accountMap  map[int64]proto.Account
	lock        *sync.RWMutex
}

func NewMaps() *Maps {
	secretMap := map[int64]proto.Secret{}
	accessesMap := map[int64]proto.Access{}
	accountMap := map[int64]proto.Account{}
	return &Maps{
		secretMap:   secretMap,
		accessesMap: accessesMap,
		accountMap:  accountMap,
		lock:        &sync.RWMutex{},
	}
}
