package kafka

import (
	"context"
	"encoding/json"
	"log"
	"order-service/config"
	"order-service/internal/cache"
	"order-service/internal/db"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	db     *db.Database
	cache  *cache.Cache
}

func NewConsumer(cfg *config.Config, db *db.Database, cache *cache.Cache) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.KafkaBrokers,
			Topic:          cfg.KafkaTopic,
			GroupID:        "order-service",
			MinBytes:       10e3,
			MaxBytes:       10e6,
			CommitInterval: time.Second,
		}),
		db:    db,
		cache: cache}
}

func (c *Consumer) start() {
	go c.ConsumeMessages()
}

func (c *Consumer) ConsumeMessages() {
	for {
		msg, err := c.reader.FetchMessage(context.Background())
		if err != nil {
			log.Printf("Failed to fetch message: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		var order db.Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			c.reader.CommitMessages(context.Background(), msg)
			continue
		}

		if err := c.db.SaveOrder(&order); err != nil {
			log.Printf("Failed to save order to DB: %v", err)
		} else {
			c.cache.Set(&order)
			log.Printf("Order %s saved and cached", order.OrderUID)
		}

		if err := c.reader.CommitMessages(context.Background(), msg); err != nil {
			log.Printf("Failed to commit message: %v", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
