package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/alexveli/astral-praktika/internal/domain"
	mylog "github.com/alexveli/astral-praktika/pkg/log"
)

func (h *Handler) GenerateSecret(c *gin.Context) {
	token := h.extractToken(c)
	account, err := h.services.Authenticator.GetAccountFromToken(c.Request.Context(), token)
	if err != nil {
		mylog.SugarLogger.Warnf("cannot get account, %v", err)
		newResponse(c, http.StatusUnauthorized, err.Error())

		return
	}
	key, isGenerated := h.services.SecretKeeper.GenerateSecret(c.Request.Context(), account)
	if !isGenerated {
		newResponse(c, http.StatusInternalServerError, "failed to generate key")

		return
	}
	newResponse(c, http.StatusOK, key)
}

func (h *Handler) GetSecret(c *gin.Context) {
	token := h.extractToken(c)
	account, err := h.services.Authenticator.GetAccountFromToken(c.Request.Context(), token)
	if err != nil {
		newResponse(c, http.StatusUnauthorized, err.Error())

		return
	}
	key, keyValid := h.services.SecretKeeper.ValidateKey(c.Param("key"))
	if !keyValid {
		newResponse(c, http.StatusUnprocessableEntity, "provided key has invalid format")

		return
	}
	secret, storedUserID, err := h.services.SecretKeeper.ProvideSecret(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, domain.ErrSecretHasExpired) {
			newResponse(c, http.StatusNotAcceptable, domain.ErrSecretHasExpired)

			return
		}
		if errors.Is(err, domain.ErrSecretAccessesCountExceeded) {
			newResponse(c, http.StatusNotAcceptable, domain.ErrSecretAccessesCountExceeded)

			return
		}
		newResponse(c, http.StatusInternalServerError, "cannot provide secret")

		return
	}
	if storedUserID != account.UserID {
		newResponse(c, http.StatusConflict, "cannot provide secret of another user")

		return
	}
	newResponse(c, http.StatusOK, secret)
}
