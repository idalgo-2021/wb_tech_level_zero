package kafkadelivery

type OrderMessage struct {
	OrderID   string  `json:"order_id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	UserID    string  `json:"user_id"`
	Status    string  `json:"status"`
}

type ValidationError struct {
	Field   string
	Message string
}
