package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
)

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

type OrderGenerator struct {
	writer *kafka.Writer
}

func NewOrderGenerator(brokers, topic string) *OrderGenerator {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &OrderGenerator{
		writer: writer,
	}
}

func (g *OrderGenerator) GenerateOrder() Order {
	orderUID := gofakeit.UUID()
	trackNumber := "WB" + gofakeit.Regex("[A-Z0-9]{10}")

	delivery := Delivery{
		Name:    gofakeit.Name(),
		Phone:   gofakeit.Phone(),
		Zip:     gofakeit.Zip(),
		City:    gofakeit.City(),
		Address: gofakeit.Street() + ", " + gofakeit.Digit(),
		Region:  gofakeit.State(),
		Email:   gofakeit.Email(),
	}

	amount := gofakeit.Number(10000, 500000)
	deliveryCost := gofakeit.Number(1000, 10000)
	goodsTotal := amount - deliveryCost
	if goodsTotal < 0 {
		goodsTotal = amount - 5000
	}

	payment := Payment{
		Transaction:  orderUID,
		RequestID:    gofakeit.UUID(),
		Currency:     "USD",
		Provider:     "wbpay",
		Amount:       amount,
		PaymentDt:    time.Now().Unix(),
		Bank:         "alpha",
		DeliveryCost: deliveryCost,
		GoodsTotal:   goodsTotal,
		CustomFee:    gofakeit.Number(0, 2000),
	}

	itemCount := rand.Intn(3) + 1
	items := make([]Item, itemCount)
	for i := 0; i < itemCount; i++ {
		price := gofakeit.Number(1000, 50000)
		sale := gofakeit.Number(0, 5000)

		items[i] = Item{
			ChrtID:      gofakeit.Number(1000000, 9999999),
			TrackNumber: trackNumber,
			Price:       price,
			Rid:         gofakeit.UUID(),
			Name:        gofakeit.ProductName(),
			Sale:        sale,
			Size:        strconv.Itoa(gofakeit.Number(1, 10)),
			TotalPrice:  price - sale,
			NmID:        gofakeit.Number(100000, 999999),
			Brand:       gofakeit.Company(),
			Status:      gofakeit.Number(100, 400),
		}
	}

	return Order{
		OrderUID:          orderUID,
		TrackNumber:       trackNumber,
		Entry:             "WBIL",
		Delivery:          delivery,
		Payment:           payment,
		Items:             items,
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test" + strconv.Itoa(gofakeit.Number(1, 1000)),
		DeliveryService:   "meest",
		Shardkey:          strconv.Itoa(gofakeit.Number(1, 10)),
		SmID:              gofakeit.Number(1, 100),
		DateCreated:       time.Now(),
		OofShard:          strconv.Itoa(gofakeit.Number(1, 10)),
	}
}

func (g *OrderGenerator) SendOrder(order Order) error {
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(order.OrderUID),
		Value: orderJSON,
	}

	err = g.writer.WriteMessages(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	amountUSD := float64(order.Payment.Amount) / 100.0
	log.Printf("Order sent successfully: %s (track: %s, amount: $%.2f)",
		order.OrderUID, order.TrackNumber, amountUSD)
	return nil
}

func (g *OrderGenerator) Run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Starting order generator. Producing orders every %v to topic: orders\n", interval)
	log.Println("Press Ctrl+C to stop...")

	for {
		select {
		case <-ticker.C:
			order := g.GenerateOrder()
			if err := g.SendOrder(order); err != nil {
				log.Printf("Error sending order: %v\n", err)
			}

		case <-sigchan:
			log.Println("Shutting down order generator...")
			g.writer.Close()
			return
		}
	}
}

func main() {
	brokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	topic := getEnv("ORDERS_TOPIC", "orders")
	intervalStr := getEnv("GENERATION_INTERVAL", "5s")

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.Printf("Invalid interval, using default 5s: %v", err)
		interval = 5 * time.Second
	}

	generator := NewOrderGenerator(brokers, topic)
	defer generator.writer.Close()

	generator.Run(interval)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
