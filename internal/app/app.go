package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexveli/astral-praktika/internal/config"
	"github.com/alexveli/astral-praktika/internal/domain"
	"github.com/alexveli/astral-praktika/internal/repository"
	"github.com/alexveli/astral-praktika/internal/service"
	"github.com/alexveli/astral-praktika/internal/transport/handlers"
	"github.com/alexveli/astral-praktika/pkg/auth"
	"github.com/alexveli/astral-praktika/pkg/hash"
	mylog "github.com/alexveli/astral-praktika/pkg/log"
	"github.com/alexveli/astral-praktika/pkg/storage/maps"
	"github.com/alexveli/astral-praktika/pkg/storage/postgres"
)

func Run() {

	mylog.SugarLogger = mylog.InitLogger("keymaster.out")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Config{}

	setFlags(&cfg)

	_, err := config.NewConfig(&cfg)
	if err != nil {
		mylog.SugarLogger.Errorf("Cannot get config: %v", err)

		_, _ = fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(2)
	}
	mylog.SugarLogger.Infof("Config set: %v", cfg)

	mylog.SugarLogger.Infof("Flags set: %v", cfg)

	var repos *repository.Repositories
	if cfg.Storage.StorageDriver == domain.POSTGRES {
		db, err := postgres.NewPostgresDB(cfg.Postgres.DatabaseURI)
		if err != nil {
			mylog.SugarLogger.Fatalf("Cannot connect to db: %v", err)
		}
		mylog.SugarLogger.Infof("DB set: %v", db)

		repos = repository.NewRepositories(db)
		mylog.SugarLogger.Infof("Repos set: %v", repos)

		dbmanager := repository.NewDBCreator(db)
		err = dbmanager.CreateTables(ctx)
		if err != nil {
			mylog.SugarLogger.Fatalf("cannot generate tables in database, %v", err)
		}
		//testing
		_, _ = db.Exec(ctx, "TRUNCATE accounts;TRUNCATE accesses;TRUNCATE secrets;")
	}
	if cfg.Storage.StorageDriver == domain.MAP {
		mapTables := maps.NewMaps()
		repos = repository.NewMapRepositories(mapTables)
	}

	tokenManager, err := auth.NewManager(cfg.Auth.JWT)
	if err != nil {
		mylog.SugarLogger.Fatalf("Cannot initiate tockenManager: %v", err)
	}
	mylog.SugarLogger.Infof("TokenManager set: %v", tokenManager)

	hasher := hash.NewHasher(cfg.Hash.Key)
	mylog.SugarLogger.Infof("Hasher set: %v", hasher)

	services := service.NewServices(repos, cfg.Keeper, tokenManager)
	mylog.SugarLogger.Infof("Services set: %v", services)

	handlerSet := handlers.NewHandler(services, hasher)
	mylog.SugarLogger.Infof("handlerSet set: %v", handlerSet)

	startServer(ctx, &cfg, handlerSet)
}

func setFlags(cfg *config.Config) {
	flag.StringVar(&cfg.Server.RunAddress, "a", cfg.Server.RunAddress, "address for starting server")
	flag.StringVar(&cfg.Postgres.DatabaseURI, "d", cfg.Postgres.DatabaseURI, "database connection string")
	flag.Parse()
	err := os.Setenv(domain.DatabaseUri, cfg.Postgres.DatabaseURI)
	if err != nil {
		mylog.SugarLogger.Infof("cannot set DATABASE_URI, %v", err)

		return
	}
	err = os.Setenv(domain.RunAddress, cfg.Server.RunAddress)
	if err != nil {
		mylog.SugarLogger.Infof("cannot set RUN_ADDRESS, %v", err)

		return
	}

}

func startServer(ctx context.Context, config *config.Config, handler *handlers.Handler) {
	srv := &http.Server{
		Addr:    config.Server.RunAddress,
		Handler: handler.Init(),
	}
	quit := make(chan os.Signal, 1)
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			mylog.SugarLogger.Errorf("error occurred while running http server: %s\n", err.Error())
			quit <- syscall.SIGTERM
		}
	}()
	// Graceful Shutdown on system interrupt
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)

	<-quit

	mylog.SugarLogger.Info("terminating")

	ctx, shutdown := context.WithTimeout(ctx, config.Server.ShutdownTimeout)
	defer shutdown()

	if err := srv.Shutdown(ctx); err != nil {
		mylog.SugarLogger.Errorf("failed to stop server: %v", err)
	}
}
