package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/abhi-jeet589/mocksmith/internal/config"
	"github.com/abhi-jeet589/mocksmith/internal/repository"
	"github.com/abhi-jeet589/mocksmith/internal/server"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repo, err := repository.New(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer func() {
		if err := repo.Close(context.Background()); err != nil {
			log.Printf("mongo disconnect: %v", err)
		}
	}()

	if err := server.New(cfg, repo).Run(ctx); err != nil {
		log.Fatalf("server: %v", err)
	}
}
