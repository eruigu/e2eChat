package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func runServer() error {
	const addr = "localhost:8080"

	// Start listening
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("listening on ws://%v", l.Addr())

	// Initialize chat server
	cs := newChatServer()

	// Create HTTP server
	s := &http.Server{
		Handler:      cs,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Serve in a goroutine
	errc := make(chan error, 1)
	go func() {
		errc <- s.Serve(l)
	}()

	// Handle Ctrl+C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	select {
	case err := <-errc:
		log.Printf("server error: %v", err)
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.Shutdown(ctx)
}
