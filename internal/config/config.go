package config

import (
	"time"

	"github.com/caarlos0/env/v6"

	mylog "github.com/alexveli/astral-praktika/pkg/log"
)

type (
	Config struct {
		Postgres PostgresConfig
		Storage  StorageConfig
		Keeper   KeeperConfig
		Server   HTTPServerConfig
		Hash     HashConfig
		Auth     AuthConfig
	}
	StorageConfig struct {
		StorageDriver string `env:"STORAGE_DRIVER" envDefault:"POSTGRES"`
	}
	PostgresConfig struct {
		DatabaseURI string        `env:"DATABASE_URI,unset"`
		Timeout     time.Duration `env:"DATABASE_TIMEOUT" envDefault:"30s"`
	}
	KeeperConfig struct {
		AccessCount      int64         `env:"ACCESS_COUNT" envDefault:"3"`
		ExpirationPeriod time.Duration `env:"EXPIRATION_PERIOD" envDefault:"72h"`
		SecretLength     int64         `env:"SECRET_LENGTH" envDefault:"500"`
		ExpireExisting   bool          `env:"EXPIRE_EXISTING" envDefault:"true"`
		KeyLength        int64         `env:"KEY_LENGTH" envDefault:"10"`
	}
	HTTPServerConfig struct {
		RunAddress      string        `env:"RUN_ADDRESS" envDefault:"127.0.0.1:8081"`
		ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"2s"`
	}
	HashConfig struct {
		Key string `env:"KEY" envDefault:"j3n4b%21&#"`
	}
	AuthConfig struct {
		JWT                    JWTConfig
		PasswordSalt           string `env:"SALT,unset" envDefault:"kjH^#(876320"`
		VerificationCodeLength int    `env:"VERIFICATION_CODE_LEN" envDefault:"8"`
	}
	JWTConfig struct {
		AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"15m"`
		RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL" envDefault:"24h"`
		SigningKey      string        `env:"SIGNING_KEY" envDefault:"Ed1039%^&*3JS"`
	}
)

func NewConfig(cfg *Config) (*Config, error) {
	mylog.SugarLogger.Infoln("Init Config")
	if err := env.Parse(cfg); err != nil {
		mylog.SugarLogger.Errorf("%+v", err)
		return nil, err
	}
	mylog.SugarLogger.Infof("%+v", cfg)
	return cfg, nil
}
