package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"

	mylog "github.com/alexveli/astral-praktika/pkg/log"
)

func (h *Handler) extractToken(c *gin.Context) string {
	token := c.Query("token")
	if token != "" {

		return token
	}
	bearerToken := c.Request.Header.Get("Authorization")
	if bearerToken != "" {
		if strings.Contains(bearerToken, " ") {
			if len(strings.Split(bearerToken, " ")) == 2 {

				return strings.Split(bearerToken, " ")[1]
			}
		}
		return bearerToken
	}
	mylog.SugarLogger.Warnf("token is empty %s", bearerToken)

	return ""
}
