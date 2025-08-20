package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"order-service/config"
	"time"

	_ "github.com/lib/pq"
)

type Database struct {
	Conn *sql.DB
}

func NewDB(cfg *config.Config) (*Database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	log.Printf("Connecting to database with: %s", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Successfully connected to database")
	return &Database{Conn: db}, nil
}

func (d *Database) Close() error {
	return d.Conn.Close()
}

func (d *Database) SaveOrder(order *Order) error {
	log.Printf("Saving order: %s", order.OrderUID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := d.Conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO UPDATE SET
			track_number = EXCLUDED.track_number,
			entry = EXCLUDED.entry,
			locale = EXCLUDED.locale,
			internal_signature = EXCLUDED.internal_signature,
			customer_id = EXCLUDED.customer_id,
			delivery_service = EXCLUDED.delivery_service,
			shardkey = EXCLUDED.shardkey,
			sm_id = EXCLUDED.sm_id,
			date_created = EXCLUDED.date_created,
			oof_shard = EXCLUDED.oof_shard`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		log.Printf("Error saving order: %v", err)
		return fmt.Errorf("failed to save order: %w", err)
	}
	rows, _ := result.RowsAffected()
	log.Printf("Orders affected: %d", rows)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO deliveries (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (order_uid) DO UPDATE SET
			name = EXCLUDED.name,
			phone = EXCLUDED.phone,
			zip = EXCLUDED.zip,
			city = EXCLUDED.city,
			address = EXCLUDED.address,
			region = EXCLUDED.region,
			email = EXCLUDED.email`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("failed to save delivery: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO payments (
			order_uid, transaction, request_id, currency, provider, amount,
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO UPDATE SET
			transaction = EXCLUDED.transaction,
			request_id = EXCLUDED.request_id,
			currency = EXCLUDED.currency,
			provider = EXCLUDED.provider,
			amount = EXCLUDED.amount,
			payment_dt = EXCLUDED.payment_dt,
			bank = EXCLUDED.bank,
			delivery_cost = EXCLUDED.delivery_cost,
			goods_total = EXCLUDED.goods_total,
			custom_fee = EXCLUDED.custom_fee`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM items WHERE order_uid = $1", order.OrderUID)
	if err != nil {
		return fmt.Errorf("failed to delete old items: %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO items (
				order_uid, chrt_id, track_number, price, rid, name, sale,
				size, total_price, nm_id, brand, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price,
			item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("failed to save item: %w", err)
		}
	}

	return tx.Commit()
}

func (d *Database) GetOrderByUID(uid string) (*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	err := d.Conn.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM orders WHERE order_uid = $1)", uid).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check order existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("order not found")
	}

	order := &Order{}

	err = d.Conn.QueryRowContext(ctx, `
		SELECT order_uid, track_number, entry, locale, internal_signature,
		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
		WHERE order_uid = $1`, uid).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	err = d.Conn.QueryRowContext(ctx, `
		SELECT name, phone, zip, city, address, region, email
		FROM deliveries
		WHERE order_uid = $1`, uid).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	err = d.Conn.QueryRowContext(ctx, `
		SELECT transaction, request_id, currency, provider, amount, payment_dt,
			bank, delivery_cost, goods_total, custom_fee
		FROM payments
		WHERE order_uid = $1`, uid).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
		&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
		&order.Payment.CustomFee)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	rows, err := d.Conn.QueryContext(ctx, `
		SELECT chrt_id, track_number, price, rid, name, sale, size,
			total_price, nm_id, brand, status
		FROM items
		WHERE order_uid = $1`, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	defer rows.Close()

	order.Items = []Item{}
	for rows.Next() {
		var item Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (d *Database) GetAllOrders() ([]Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := d.Conn.QueryContext(ctx, "SELECT order_uid FROM orders")
	if err != nil {
		if err == sql.ErrNoRows {
			return []Order{}, nil
		}
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("failed to scan order uid: %w", err)
		}

		order, err := d.GetOrderByUID(uid)
		if err != nil {
			log.Printf("Failed to load order %s: %v", uid, err)
			continue
		}
		orders = append(orders, *order)
	}

	if len(orders) == 0 {
		return []Order{}, nil
	}

	return orders, nil
}

func (d *Database) LoadTestData() error {
	testData := `{
		"order_uid": "b563feb7b2b84b6test",
		"track_number": "WBILMTESTTRACK",
		"entry": "WBIL",
		"delivery": {
			"name": "Test Testov",
			"phone": "+9720000000",
			"zip": "2639809",
			"city": "Kiryat Mozkin",
			"address": "Ploshad Mira 15",
			"region": "Kraiot",
			"email": "test@gmail.com"
		},
		"payment": {
			"transaction": "b563feb7b2b84b6test",
			"request_id": "",
			"currency": "USD",
			"provider": "wbpay",
			"amount": 1817,
			"payment_dt": 1637907727,
			"bank": "alpha",
			"delivery_cost": 1500,
			"goods_total": 317,
			"custom_fee": 0
		},
		"items": [
			{
				"chrt_id": 9934930,
				"track_number": "WBILMTESTTRACK",
				"price": 453,
				"rid": "ab4219087a764ae0btest",
				"name": "Mascaras",
				"sale": 30,
				"size": "0",
				"total_price": 317,
				"nm_id": 2389212,
				"brand": "Vivienne Sabo",
				"status": 202
			}
		],
		"locale": "en",
		"internal_signature": "",
		"customer_id": "test",
		"delivery_service": "meest",
		"shardkey": "9",
		"sm_id": 99,
		"date_created": "2021-11-26T06:22:19Z",
		"oof_shard": "1"
	}`

	var order Order
	if err := json.Unmarshal([]byte(testData), &order); err != nil {
		return fmt.Errorf("failed to unmarshal test data: %w", err)
	}

	return d.SaveOrder(&order)
}

func (d *Database) OrderExists(uid string) (bool, error) {
	var exists bool
	err := d.Conn.QueryRow("SELECT EXISTS(SELECT 1 FROM orders WHERE order_uid = $1)", uid).Scan(&exists)
	return exists, err
}
