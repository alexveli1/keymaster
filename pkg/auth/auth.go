package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/alexveli/astral-praktika/internal/config"
	"github.com/alexveli/astral-praktika/internal/domain"
	"github.com/alexveli/astral-praktika/internal/proto"
	mylog "github.com/alexveli/astral-praktika/pkg/log"
)

type Manager struct {
	signingKey      string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewManager(cfg config.JWTConfig) (*Manager, error) {
	if cfg == (config.JWTConfig{}) {
		return nil, errors.New("empty jwt configuration provided")
	}
	return &Manager{
		signingKey:      cfg.SigningKey,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}, nil
}

type Claims struct {
	jwt.StandardClaims
	Username string `json:"username"`
}

func (m *Manager) GenerateToken(userID int64, tokenType string) (*proto.TokenAndExpiresAt, error) {
	var expiresAt int64
	if tokenType == domain.ACCESS {
		expiresAt = time.Now().Add(m.accessTokenTTL).Unix()
	} else {
		expiresAt = time.Now().Add(m.refreshTokenTTL).Unix()
	}
	claims := jwt.StandardClaims{
		ExpiresAt: expiresAt,
		Subject:   fmt.Sprint(userID),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	generatedToken, err := token.SignedString([]byte(m.signingKey))

	return &proto.TokenAndExpiresAt{
		Uuid:      userID,
		Token:     generatedToken,
		ExpiresAt: timestamppb.New(time.Unix(expiresAt, 0)),
	}, err
}

func (m *Manager) TokenValid(token string) error {
	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			mylog.SugarLogger.Errorf("unexpected signing method: %v", token.Header["alg"])

			return nil, domain.ErrAutorizationSigningMethod
		}

		return []byte(m.signingKey), nil
	})
	if err != nil {
		mylog.SugarLogger.Errorf("cannot parse token %v, %v", token, err)

		return err
	}

	return nil
}

func (m *Manager) ExtractUserIDFromToken(token string) (int64, error) {
	if token == "" {

		return 0, domain.ErrAuthorizationInvalidToken
	}
	claims := jwt.StandardClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {

			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(m.signingKey), nil
	})
	if err != nil {
		return 0, err
	}
	if parsedToken.Valid {
		userid, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			mylog.SugarLogger.Errorf("cannot convert userid to int64, %v", err)

			return 0, err
		}

		return userid, nil
	}
	mylog.SugarLogger.Infof("parsedToken not valid")
	return 0, domain.ErrAuthorizationInvalidToken
}
