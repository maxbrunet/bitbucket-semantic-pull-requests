// Package bitbucket-semantic-pull-requests is the entrypoint of the webhook server
package main

import (
	goflag "flag"
	"log"
	"net/http"
	"os"
	"time"

	flag "github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/maxbrunet/bitbucket-semantic-pull-requests/internal/handler"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	bitbucketUsername := flag.String(
		"bitbucket-username",
		"",
		"Bitbucket username (env BITBUCKET_USERNAME)",
	)
	bitbucketPassword := flag.String(
		"bitbucket-password",
		"",
		"Bitbucket password (env BITBUCKET_PASSWORD)",
	)
	listenAddr := flag.String("listen-address", ":8888", "Address to listen on for the webhook")
	logLevel := zap.LevelFlag(
		"log-level",
		zap.InfoLevel,
		"Only log messages with the given severity or above. One of: [debug, info, warn, error, dpanic, panic, fatal]",
	)

	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()

	if env, ok := os.LookupEnv("BITBUCKET_USERNAME"); ok && *bitbucketUsername == "" {
		*bitbucketUsername = env
	}

	if env, ok := os.LookupEnv("BITBUCKET_PASSWORD"); ok && *bitbucketPassword == "" {
		*bitbucketPassword = env
	}

	loggerCfg := zap.NewProductionConfig()
	loggerCfg.Level = zap.NewAtomicLevelAt(*logLevel)

	logger, err := loggerCfg.Build()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info(
		"starting semantic-pull-requests",
		zap.String("version", version),
		zap.String("commit", commit),
		zap.String("date", date),
		zap.String("listen-address", *listenAddr),
	)

	if *bitbucketUsername == "" || *bitbucketPassword == "" {
		logger.Fatal("Bitbucket username and password are both required")
	}

	spr, err := handler.NewSemanticPullRequests(*bitbucketUsername, *bitbucketPassword, logger)
	if err != nil {
		logger.Fatal("failed to initialize semantic-pull-requests", zap.Error(err))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", spr.HandlePullRequestUpdate)

	errorLog, err := zap.NewStdLogAt(logger, zap.ErrorLevel)
	if err != nil {
		logger.Fatal(
			"failed to create standard error logger",
			zap.Error(err),
		)
	}

	//nolint:mnd
	server := &http.Server{
		Addr:              *listenAddr,
		Handler:           mux,
		ErrorLog:          errorLog,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("error starting HTTP server", zap.Error(err))
	}
}
