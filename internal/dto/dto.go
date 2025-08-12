package dto

import (
	"time"
	"wb_tech_level_zero/internal/orders"
)

type OrderDTO struct {
	// ID          uuid.UUID `json:"id"`
	OrderUID    string `json:"order_uid"`
	TrackNumber string `json:"track_number"`
	Entry       string `json:"entry"`

	Delivery DeliveryDTO `json:"delivery"`
	Payment  PaymentDTO  `json:"payment"`
	Items    []ItemDTO   `json:"items"`

	Locale            string     `json:"locale"`
	InternalSignature string     `json:"internal_signature"`
	CustomerID        string     `json:"customer_id"`
	DeliveryService   string     `json:"delivery_service"`
	Shardkey          string     `json:"shardkey"`
	SmID              int        `json:"sm_id"`
	DateCreated       *time.Time `json:"date_created"`
	OofShard          string     `json:"oof_shard"`
}

type DeliveryDTO struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type PaymentDTO struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type ItemDTO struct {
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

type OrdersResponse struct {
	Orders []OrderDTO `json:"orders"`
	Total  int        `json:"total"`
	Page   int        `json:"page"`
	Limit  int        `json:"limit"`
}

type ErrorResponse struct {
	Message string `json:"message" example:"An unexpected error occurred."`
}

///////////////////

func ItemsToDTO(items []orders.Item) []ItemDTO {
	result := make([]ItemDTO, 0, len(items))
	for _, i := range items {
		result = append(result, ItemDTO{
			ChrtID:      i.ChrtID,
			TrackNumber: i.TrackNumber,
			Price:       i.Price,
			Rid:         i.Rid,
			Name:        i.Name,
			Sale:        i.Sale,
			Size:        i.Size,
			TotalPrice:  i.TotalPrice,
			NmID:        i.NmID,
			Brand:       i.Brand,
			Status:      i.Status,
		})
	}
	return result
}
func OrderToDTO(o *orders.Order) OrderDTO {
	return OrderDTO{
		OrderUID:          o.OrderUID,
		TrackNumber:       o.TrackNumber,
		Entry:             o.Entry,
		Delivery:          DeliveryDTO(o.Delivery),
		Payment:           PaymentDTO(o.Payment),
		Items:             ItemsToDTO(o.Items),
		Locale:            o.Locale,
		InternalSignature: o.InternalSignature,
		CustomerID:        o.CustomerID,
		DeliveryService:   o.DeliveryService,
		Shardkey:          o.Shardkey,
		SmID:              o.SmID,
		DateCreated:       o.DateCreated,
		OofShard:          o.OofShard,
	}
}
