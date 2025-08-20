package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"order-service/internal/cache"
	"order-service/internal/db"
)

type OrderHandler struct {
	cache *cache.Cache
	db    *db.Database
}

func NewOrderHandler(cache *cache.Cache, db *db.Database) *OrderHandler {
	return &OrderHandler{cache, db}
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Path[len("/order/"):]
	if uid == "" {
		http.Error(w, "order UID is required", http.StatusBadRequest)
		return
	}

	if order, ok := h.cache.Get(uid); ok {
		respondWithJSON(w, http.StatusOK, order)
		return
	}

	order, err := h.db.GetOrderByUID(uid)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	h.cache.Set(order)
	respondWithJSON(w, http.StatusOK, order)
}

func respondWithJSON(w http.ResponseWriter, StatusCode int, data interface{}) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(StatusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "web/static/index.html")
}
