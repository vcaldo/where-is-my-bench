package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/internal/config"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/internal/handlers"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/pkg/telegram"
)

const nrShutdownTimeout = 5 * time.Second

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Initialize NewRelic if license key is provided
	nrApp, err := initNewRelic(cfg)
	if err != nil {
		log.Printf("failed to initialize New Relic, continuing without APM: %v", err)
		nrApp = nil
	}

	mainTxn := nrApp.StartTransaction("bot_startup")
	mainTxn.AddAttribute("environment", cfg.Environment)
	defer mainTxn.End()

	ctx = newrelic.NewContext(ctx, mainTxn)

	opts := []bot.Option{
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			updateTxn := nrApp.StartTransaction("bot_update")
			updateTxn.AddAttribute("update_type", "message")
			updateCtx := newrelic.NewContext(ctx, updateTxn)
			defer updateTxn.End()
			handlers.Handler(updateCtx, b, update)
		}),
	}

	segment := mainTxn.StartSegment("bot_init")
	b, err := telegram.NewClient(cfg.TelegramToken, opts...)
	if err != nil {
		mainTxn.NoticeError(err)
		log.Fatalf("error creating telegram client: %v", err)
	}
	segment.End()

	// Start bot in goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Println("Bot started successfully")
		b.Start(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Println("Shutdown signal received, initiating graceful shutdown...")
	case err := <-errChan:
		log.Printf("Bot error: %v, initiating shutdown...", err)
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), nrShutdownTimeout)
	defer shutdownCancel()

	b.Close(shutdownCtx)
	log.Println("Bot shutdown initiated")
	log.Println("Bot shutdown complete")

	if nrApp != nil {
		nrApp.Shutdown(nrShutdownTimeout)
		log.Println("NewRelic shutdown complete")
	}
}

func initNewRelic(cfg *config.Config) (*newrelic.Application, error) {
	if cfg.NewRelicLicenseKey == "" {
		return nil, fmt.Errorf("new relic license key is not set")
	}

	return newrelic.NewApplication(
		newrelic.ConfigAppName(cfg.NewRelicAppName),
		newrelic.ConfigLicense(cfg.NewRelicLicenseKey),
		newrelic.ConfigDistributedTracerEnabled(true),
		newrelic.ConfigAppLogEnabled(true),
		// newrelic.ConfigDebugLogger(os.Stdout),
	)
}
