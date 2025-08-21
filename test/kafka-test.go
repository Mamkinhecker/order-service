package test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"order-service/internal/db"
	"time"

	"github.com/segmentio/kafka-go"
)

func Test() {
	w := &kafka.Writer{
		Addr:     kafka.TCP("redpanda:9092"),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}

	defer w.Close()

	order := db.Order{
		OrderUID:    "test-order-123",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: db.Delivery{
			Name:    "YURY",
			Phone:   "8033412412",
			Zip:     "220001",
			City:    "Toronto",
			Address: "test_adress",
			Region:  "slavia",
			Email:   "YURYkrutoi123@gmail.com",
		},
		Payment: db.Payment{
			Transaction:  "b2kbm2kn2kbn2k",
			RequestID:    "",
			Currency:     "EURO",
			Provider:     "wbpay",
			Amount:       9,
			PaymentDt:    813812,
			Bank:         "PRIORbank",
			DeliveryCost: 12,
			GoodsTotal:   1232,
			CustomFee:    19329,
		},
		Items: []db.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "cars",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "BelGee",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "test",
		DeliveryService: "meest",
		Shardkey:        "9",
		SmID:            99,
		DateCreated:     time.Now(),
		OofShard:        "1",
	}

	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Fatal("Failed to marshal order:", err)
	}

	err = w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(order.OrderUID),
			Value: orderJSON,
		},
	)
	if err != nil {
		log.Fatal("Failed to write message:", err)
	}

	fmt.Println("Message sent successfully:", string(orderJSON))
}
