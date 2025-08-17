package service

import (
	"wb_tech_level_zero/internal/delivery/kafkadelivery"
	"wb_tech_level_zero/internal/orders"
)

type GetOrdersParams struct {
	Page  int
	Limit int
}

func mapEventOrderToDomain(eo *kafkadelivery.EventOrder) orders.Order {
	order := orders.Order{
		OrderUID:          eo.OrderUID,
		TrackNumber:       eo.TrackNumber,
		Entry:             eo.Entry,
		Locale:            eo.Locale,
		InternalSignature: eo.InternalSignature,
		CustomerID:        eo.CustomerID,
		DeliveryService:   eo.DeliveryService,
		Shardkey:          eo.Shardkey,
		SmID:              eo.SmID,
		DateCreated:       &eo.DateCreated,
		OofShard:          eo.OofShard,
		Delivery: orders.Delivery{
			Name:    eo.Delivery.Name,
			Phone:   eo.Delivery.Phone,
			Zip:     eo.Delivery.Zip,
			City:    eo.Delivery.City,
			Address: eo.Delivery.Address,
			Region:  eo.Delivery.Region,
			Email:   eo.Delivery.Email,
		},
		Payment: orders.Payment{
			Transaction:  eo.Payment.Transaction,
			RequestID:    eo.Payment.RequestID,
			Currency:     eo.Payment.Currency,
			Provider:     eo.Payment.Provider,
			Amount:       eo.Payment.Amount,
			PaymentDT:    eo.Payment.PaymentDT,
			Bank:         eo.Payment.Bank,
			DeliveryCost: eo.Payment.DeliveryCost,
			GoodsTotal:   eo.Payment.GoodsTotal,
			CustomFee:    eo.Payment.CustomFee,
		},
	}

	items := make([]orders.Item, len(eo.Items))
	for i, it := range eo.Items {
		items[i] = orders.Item{
			ChrtID: it.ChrtID, TrackNumber: it.TrackNumber, Price: it.Price,
			Rid: it.Rid, Name: it.Name, Sale: it.Sale, Size: it.Size,
			TotalPrice: it.TotalPrice, NmID: it.NmID, Brand: it.Brand, Status: it.Status,
		}
	}
	order.Items = items

	return order
}
