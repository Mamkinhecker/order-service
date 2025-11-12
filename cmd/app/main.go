package main

import (
	"log"
	"net/http"
	"order-service/config"
	"order-service/internal/cache"
	"order-service/internal/db"
	"order-service/internal/handlers"
	"order-service/internal/kafka"
	"order-service/test"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.Load()

	database, err := db.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	exists, err := database.OrderExists("b563feb7b2b84b6test")
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		if err := database.LoadTestData(); err != nil {
			log.Fatal(err)
		}
		log.Println("Test data loaded successfully")
	} else {
		log.Println("Test data already exists")
	}

	c := cache.NewCache()
	orders, err := database.GetAllOrders()
	if err != nil {
		log.Printf("Failed to restore cache from DB: %v", err)
	} else {
		c.Restore(orders)
		log.Printf("Restored %d orders to cache", len(orders))
	}

	consumer := kafka.NewConsumer(cfg, database, c)
	consumer.Start()
	defer consumer.Close()

	orderHandler := handlers.NewOrderHandler(c, database)
	http.HandleFunc("/order/", orderHandler.GetOrder)
	http.HandleFunc("/", handlers.StaticHandler)

	go func() {
		log.Printf("Starting HTTP server on :%s", cfg.HTTPPort)
		if err := http.ListenAndServe(":"+cfg.HTTPPort, nil); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	test.Test()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting dowsn...")
}
