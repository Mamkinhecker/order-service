package db

import (
	"encoding/json"
	"fmt"
	"time"
)

type Order struct {
	OrderUID          string    `json:"order_uid" validate:"required"`
	TrackNumber       string    `json:"track_number" validate:"required"`
	Entry             string    `json:"entry" validate:"required,alpha,min=1,max=10"`
	Delivery          Delivery  `json:"delivery" validate:"required"`
	Payment           Payment   `json:"payment" validate:"required"`
	Items             []Item    `json:"items" validate:"required,min=1,dive"`
	Locale            string    `json:"locale" validate:"required,oneof=en ru pt"`
	InternalSignature string    `json:"internal_signature" validate:"max=100"`
	CustomerID        string    `json:"customer_id" validate:"required,alphanum,min=1,max=50"`
	DeliveryService   string    `json:"delivery_service" validate:"required,alpha,min=1,max=50"`
	Shardkey          string    `json:"shardkey" validate:"required,numeric,min=1,max=10"`
	SmID              int       `json:"sm_id" validate:"min=0,max=1000"`
	DateCreated       time.Time `json:"date_created" validate:"required"`
	OofShard          string    `json:"oof_shard" validate:"required,numeric,min=1,max=10"`
}

func (o *Order) UnmarshalJSON(data []byte) error {
	type Alias Order
	aux := &struct {
		DateCreated string `json:"date_created"`
		*Alias
	}{
		Alias: (*Alias)(o),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	parsedTime, err := time.Parse(time.RFC3339, aux.DateCreated)
	if err != nil {
		return fmt.Errorf("failed to parse date_created: %w", err)
	}

	o.DateCreated = parsedTime
	return nil
}

type Delivery struct {
	Name    string `json:"name" validate:"required,min=1,max=100"`
	Phone   string `json:"phone" validate:"required,numeric"`
	Zip     string `json:"zip" validate:"required,numeric,min=1,max=20"`
	City    string `json:"city" validate:"required,min=1,max=100"`
	Address string `json:"address" validate:"required,min=1,max=200"`
	Region  string `json:"region" validate:"required,min=1,max=100"`
	Email   string `json:"email" validate:"required,email,max=100"`
}

type Payment struct {
	Transaction  string `json:"transaction" validate:"required"`
	RequestID    string `json:"request_id" validate:"max=50"`
	Currency     string `json:"currency" validate:"required,alpha,min=3"`
	Provider     string `json:"provider" validate:"required,alpha,min=1,max=50"`
	Amount       int    `json:"amount" validate:"required,min=0"`
	PaymentDt    int64  `json:"payment_dt" validate:"required,min=1"`
	Bank         string `json:"bank" validate:"required,alpha,min=1,max=100"`
	DeliveryCost int    `json:"delivery_cost" validate:"min=0"`
	GoodsTotal   int    `json:"goods_total" validate:"min=0"`
	CustomFee    int    `json:"custom_fee" validate:"min=0"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id" validate:"required,min=1"`
	TrackNumber string `json:"track_number" validate:"required"`
	Price       int    `json:"price" validate:"required,min=0"`
	Rid         string `json:"rid" validate:"required,min=1,max=50"`
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Sale        int    `json:"sale" validate:"min=0"`
	Size        string `json:"size" validate:"max=10"`
	TotalPrice  int    `json:"total_price" validate:"min=0"`
	NmID        int    `json:"nm_id" validate:"min=0"`
	Brand       string `json:"brand" validate:"required,min=1,max=100"`
	Status      int    `json:"status" validate:"min=0,max=999"`
}
