package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := h.extractToken(c)
		err := h.services.Authenticator.TokenValid(token)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)

			return
		}
		c.Next()
	}
}
