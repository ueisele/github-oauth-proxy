package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github-oauth-proxy/pkg/proxy"
)

func main() {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		log.Fatal("$PORT must be set")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal("$PORT must be an integer")
	}

	clientId := os.Getenv("CLIENT_ID")
	if clientId == "" {
		log.Fatal("$CLIENT_ID must be set")
	}

	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatal("$CLIENT_SECRET must be set")
	}

	allowOrigin := os.Getenv("ALLOW_ORIGIN")
	if allowOrigin == "" {
		log.Fatal("$ALLOW_ORIGIN must be set")
	}

	done := make(chan error, 1)
	defer close(done)
	server := proxy.NewProxy(proxy.Config{
		Port: port,
		ClientId: clientId,
		ClientSecret: clientSecret,
		AllowOrigin: allowOrigin,
	}, done)
	server.Run()

	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, os.Interrupt, os.Kill)
	select {
	case <-quit:
		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 5 seconds.
		log.Println("Shutdown Server ...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
		log.Println("Server exiting")
	case err := <-done:
		log.Printf("listen: %s\n", err)
	}
}