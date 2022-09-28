package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Init() *gin.Engine {
	router := gin.Default()

	router.Use(
		gin.Recovery(),
		gin.Logger(),
		corsMiddleware,
	)

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	h.initAPI(router)

	return router
}

func (h *Handler) initAPI(router *gin.Engine) {
	user := router.Group("/api/user")
	{
		user.POST("/register", h.UserRegister)
		user.POST("/login", h.UserLogin)
		user.POST("/refresh", h.RefreshTokens)
	}
	secret := user.Group("/secret", h.JwtAuthMiddleware())
	{
		secret.GET("/", h.GenerateSecret)
		secret.GET("/:key", h.GetSecret)
	}
}

func corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")

	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}
