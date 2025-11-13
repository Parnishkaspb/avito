package main

import (
	"context"
	"github.com/Parnishkaspb/avito/internal/config"
	"github.com/Parnishkaspb/avito/internal/database"
	myserver "github.com/Parnishkaspb/avito/internal/server"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		db := database.New(cfg.Postgre)
		if err := db.RunDatabase(ctx); err != nil {
			log.Printf("Database error: %v", err)
			cancel()
		}
	}()

	time.Sleep(2 * time.Second)

	wg.Add(1)
	go func() {
		defer wg.Done()
		server := myserver.New(cfg.Server)
		if err := server.RunServer(ctx); err != nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Получен сигнал на выключение...")

	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Все компонены были отключены")
	case <-time.After(10 * time.Second):
		log.Println("Выключение по времени")
	}
}
