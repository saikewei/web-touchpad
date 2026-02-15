package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"web-touchpad/internal/hid"
	"web-touchpad/internal/server"
)

func main() {
	dev, err := hid.Init()
	if err != nil {
		log.Fatalf("hid init failed: %v", err)
	}
	defer dev.Close()

	srv := server.New(dev)

	go func() {
		if err := srv.Run(":8080"); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("signal received, shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)
	// dev.Close() 会在 defer 中执行
}
