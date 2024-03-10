package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/garaekz/gonvelope/internal/auth"
	"github.com/garaekz/gonvelope/internal/config"
	"github.com/garaekz/gonvelope/internal/errors"
	"github.com/garaekz/gonvelope/internal/healthcheck"
	"github.com/garaekz/gonvelope/internal/oauth"
	"github.com/garaekz/gonvelope/pkg/accesslog"
	"github.com/garaekz/gonvelope/pkg/dbcontext"
	"github.com/garaekz/gonvelope/pkg/log"
	routing "github.com/garaekz/ozzo-routing"
	"github.com/garaekz/ozzo-routing/content"
	"github.com/garaekz/ozzo-routing/cors"
	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

// Version indicates the current version of the application.
var Version = "1.0.0"

var flagConfig = flag.String("config", "./config/local.yml", "path to the config file")

func main() {
	flag.Parse()
	// create root logger tagged with server version
	logger := log.New().With(context.TODO(), "version", Version)

	// load application configurations
	cfg, err := config.Load(*flagConfig, logger, &config.OSFileSystem{})
	if err != nil {
		logger.Errorf("failed to load application configuration: %s", err)
		os.Exit(-1)
	}

	// connect to the database
	db, err := dbx.MustOpen("postgres", cfg.DSN)
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}
	db.QueryLogFunc = logDBQuery(logger)
	db.ExecLogFunc = logDBExec(logger)
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error(err)
		}
	}()

	// build HTTP server
	address := fmt.Sprintf(":%v", cfg.ServerPort)
	hs := &http.Server{
		Addr:    address,
		Handler: buildHandler(logger, dbcontext.New(db), cfg),
	}

	// start the HTTP server with graceful shutdown
	go routing.GracefulShutdown(hs, 10*time.Second, logger.Infof)
	logger.Infof("server %v is running at %v", Version, address)
	if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(err)
		os.Exit(-1)
	}
}

// buildHandler sets up the HTTP routing and builds an HTTP handler.
func buildHandler(logger log.Logger, db *dbcontext.DB, cfg *config.Config) http.Handler {
	router := routing.New()

	router.Use(
		accesslog.Handler(logger),
		errors.Handler(logger),
		content.TypeNegotiator(content.JSON),
		cors.Handler(cors.AllowAll),
	)

	healthcheck.RegisterHandlers(router, Version)

	rg := router.Group("/api/v1/")

	auth.RegisterHandlers(rg.Group(""),
		auth.NewService(auth.NewRepository(db, logger), cfg.JWTSigningKey, cfg.JWTExpiration, logger),
		logger,
	)

	authJWTHandler := auth.JWTHandler(cfg.JWTSigningKey)
	var providerConfig = oauth.ProviderConfigs{
		Google:  cfg.GoogleOAuthConfig,
		Outlook: cfg.OutlookOAuthConfig,
	}
	store := sessions.NewCookieStore([]byte(cfg.JWTSigningKey))
	oauth.RegisterHandlers(
		router.Group("/oauth2/"),
		oauth.NewService(
			oauth.NewRepository(db, logger),
			logger,
			&providerConfig,
			cfg.JWTSigningKey,
		),
		authJWTHandler,
		logger,
		store,
	)

	return router
}

// logDBQuery returns a logging function that can be used to log SQL queries.
func logDBQuery(logger log.Logger) dbx.QueryLogFunc {
	return func(ctx context.Context, t time.Duration, sql string, _ *sql.Rows, err error) {
		if err == nil {
			logger.With(ctx, "duration", t.Milliseconds(), "sql", sql).Info("DB query successful")
		} else {
			logger.With(ctx, "sql", sql).Errorf("DB query error: %v", err)
		}
	}
}

// logDBExec returns a logging function that can be used to log SQL executions.
func logDBExec(logger log.Logger) dbx.ExecLogFunc {
	return func(ctx context.Context, t time.Duration, sql string, _ sql.Result, err error) {
		if err == nil {
			logger.With(ctx, "duration", t.Milliseconds(), "sql", sql).Info("DB execution successful")
		} else {
			logger.With(ctx, "sql", sql).Errorf("DB execution error: %v", err)
		}
	}
}
