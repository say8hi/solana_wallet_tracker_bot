package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/say8hi/walletTracker/config"
	"github.com/say8hi/walletTracker/internal/infrastructure/redis"
	"github.com/say8hi/walletTracker/internal/infrastructure/solana"
	"github.com/say8hi/walletTracker/internal/service"
	"github.com/say8hi/walletTracker/pkg/logger"
	"github.com/say8hi/walletTracker/pkg/metrics"
)


func setupMetricsServer(cfg *config.Config) {

    mux := http.NewServeMux()
    
    mux.Handle("/metrics", promhttp.Handler())
    
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    go func() {
        addr := fmt.Sprintf(":%d", cfg.Monitoring.MetricsPort)
        log.Printf("Starting metrics server on %s", addr)
        if err := http.ListenAndServe(addr, mux); err != nil {
            log.Fatalf("Failed to start metrics server: %v", err)
        }
    }()
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}


  setupMetricsServer(cfg)

	log := logger.NewLogger(cfg.App.LogLevel)
	metrics := metrics.NewMetrics(cfg.App.Name)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	solanaClient, err := solana.NewSolanaWSClient(cfg.SolanaRPC.WebsocketURL, cfg.SolanaRPC.Commitment, metrics)
	if err != nil {
		log.Fatal("Failed to create Solana client", "error", err)
	}
  redisURL := fmt.Sprintf("redis://:%s@%s:%s/1", cfg.Redis.PASSWORD, cfg.Redis.HOST, cfg.Redis.PORT)

	redisClient, err := redis.NewRedisClient(
    redisURL,
		cfg.Redis.TransactionsChannel,
		cfg.Redis.CommandsChannel,
		metrics,
	)
	if err != nil {
		log.Fatal("Failed to create Redis client", "error", err)
	}

	walletManager := service.NewWalletManager(solanaClient,
    redisClient, log, metrics)

	cmdChan, err := redisClient.SubscribeToCommands(ctx)
	if err != nil {
		log.Fatal("Failed to subscribe to commands", "error", err)
	}

	go walletManager.ListenForCommands(ctx, cmdChan)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Info("Shutting down...")

	if err := walletManager.Shutdown(ctx); err != nil {
		log.Error("Error during shutdown", "error", err)
	}
}
