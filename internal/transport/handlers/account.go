package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/alexveli/astral-praktika/internal/domain"
	"github.com/alexveli/astral-praktika/internal/proto"
	mylog "github.com/alexveli/astral-praktika/pkg/log"
)

func (h *Handler) UserRegister(c *gin.Context) {
	var input proto.InputAccount
	if err := c.BindJSON(&input); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")

		return
	}
	account := proto.Account{
		Username:     input.Login,
		PasswordHash: h.hasher.GetHash(input.Password),
	}
	err := h.services.Authenticator.Register(
		c.Request.Context(),
		&account,
	)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			newResponse(c, http.StatusConflict, err.Error())

			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}
	tokens, accessToken, err := h.services.Authenticator.GenerateTokens(c.Request.Context(), account.UserID)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}
	c.Header("Authorization", accessToken)
	newResponse(c, http.StatusOK, tokens)
}

func (h *Handler) UserLogin(c *gin.Context) {
	var input proto.InputAccount
	if err := c.BindJSON(&input); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")

		return
	}
	login := proto.Account{
		Username:     input.Login,
		PasswordHash: h.hasher.GetHash(input.Password),
	}
	account, err := h.services.Authenticator.Login(
		c.Request.Context(),
		&login,
	)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			newResponse(c, http.StatusUnauthorized, err.Error())

			return
		}
		if errors.Is(err, domain.ErrPasswordIncorrect) {
			newResponse(c, http.StatusUnauthorized, err.Error())

			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}
	tokens, accessToken, err := h.services.Authenticator.GenerateTokens(c.Request.Context(), account.UserID)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot generate tokens for user %d, %v", account.UserID, err)
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}
	c.Header("Authorization", accessToken)
	newResponse(c, http.StatusOK, tokens)
}

func (h *Handler) RefreshTokens(c *gin.Context) {
	var input proto.InputRefresh
	if err := c.BindJSON(&input); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")

		return
	}
	token := input.Token
	if token == "" {
		newResponse(c, http.StatusUnauthorized, "no token provided")

		return
	}
	err := h.services.Authenticator.TokenValid(token)
	if err != nil {
		newResponse(c, http.StatusUnauthorized, "token invalid")

		return
	}
	tokens, accessToken, err := h.services.Authenticator.RefreshToken(c.Request.Context(), token)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "cannot refresh token")

		return
	}
	c.Header("Authorization", accessToken)
	newResponse(c, http.StatusOK, tokens)
}
