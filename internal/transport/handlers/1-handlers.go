package handlers

import (
	"github.com/alexveli/astral-praktika/internal/service"

	"github.com/alexveli/astral-praktika/pkg/hash"
)

type Handler struct {
	services *service.Services
	hasher   *hash.Hasher
}

func NewHandler(services *service.Services, hasher *hash.Hasher) *Handler {
	return &Handler{
		services: services,
		hasher:   hasher,
	}
}
