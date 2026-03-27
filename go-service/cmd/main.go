package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"agro-subsidy/go-service/internal/config"
	"agro-subsidy/go-service/internal/kafka"
	"agro-subsidy/go-service/internal/service"
	"agro-subsidy/go-service/internal/worker"
)

func main() {
	cfg := config.Load()
	log.Printf("starting go-service workers=%d queue=%d ml=%s",
		cfg.WorkerCount, cfg.QueueSize, cfg.MLServiceURL)

	mlClient := service.NewMLClient(cfg.MLServiceURL)
	pool := worker.NewPool(cfg.WorkerCount, cfg.QueueSize, mlClient)

	ctx, cancel := context.WithCancel(context.Background())

	// Start worker pool before Kafka so goroutines are ready to receive.
	pool.Start(ctx)

	orch := kafka.NewOrchestrator(cfg, pool)

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-quit
		log.Printf("received %s — shutting down", sig)
		cancel()         // stop consumers and workers
		pool.Stop()      // drain queue, close results channel
		orch.Close()     // close Kafka connections
	}()

	log.Println("go-service ready")
	if err := orch.Run(ctx); err != nil {
		log.Fatalf("orchestrator fatal: %v", err)
	}
	log.Println("go-service stopped")
}
