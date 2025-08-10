// repository.go
package orders

import (
	"context"
	"errors"
	"wb_tech_level_zero/internal/models"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgOrdersRepository struct {
	db *pgxpool.Pool
}

func NewPgOrdersRepository(db *pgxpool.Pool) *PgOrdersRepository {
	return &PgOrdersRepository{db: db}
}

func (r *PgOrdersRepository) GetOrders(ctx context.Context, limit, offset int) ([]*models.Order, int, error) {

	const countQuery = `SELECT COUNT(*) FROM orders;`
	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT 
			o.id, o.order_uid, o.track_number, o.entry, 
			d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
			p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, 
			p.bank, p.delivery_cost, p.goods_total, p.custom_fee
		FROM orders o
		JOIN deliveries d ON o.id = d.order_id
		JOIN payments p ON o.id = p.order_id
		ORDER BY o.date_created DESC
		LIMIT $1 OFFSET $2;
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.ID, &o.OrderUID, &o.TrackNumber, &o.Entry,
			&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City,
			&o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email,
			&o.Payment.Transaction, &o.Payment.RequestID, &o.Payment.Currency, &o.Payment.Provider,
			&o.Payment.Amount, &o.Payment.PaymentDT, &o.Payment.Bank, &o.Payment.DeliveryCost,
			&o.Payment.GoodsTotal, &o.Payment.CustomFee,
		); err != nil {
			return nil, 0, err
		}
		orders = append(orders, &o)
	}

	return orders, total, nil
}

func (r *PgOrdersRepository) GetOrderByUID(ctx context.Context, orderUID string) (*models.Order, error) {
	const query = `
		SELECT 
			o.id, o.order_uid, o.track_number, o.entry, 
			d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
			p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, 
			p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
			o.locale, o.internal_signature, o.customer_id, o.delivery_service,
			o.shardkey, o.sm_id, o.date_created, o.oof_shard
		FROM orders o
		JOIN deliveries d ON o.id = d.order_id
		JOIN payments p ON o.id = p.order_id
		WHERE o.order_uid = $1;
	`

	var o models.Order
	err := r.db.QueryRow(ctx, query, orderUID).Scan(
		&o.ID, &o.OrderUID, &o.TrackNumber, &o.Entry,
		&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City,
		&o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email,
		&o.Payment.Transaction, &o.Payment.RequestID, &o.Payment.Currency, &o.Payment.Provider,
		&o.Payment.Amount, &o.Payment.PaymentDT, &o.Payment.Bank, &o.Payment.DeliveryCost,
		&o.Payment.GoodsTotal, &o.Payment.CustomFee,
		&o.Locale, &o.InternalSignature, &o.CustomerID, &o.DeliveryService,
		&o.Shardkey, &o.SmID, &o.DateCreated, &o.OofShard,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrOrderNotFound
		}
		return nil, err
	}

	// Получаем items
	const itemsQuery = `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items
		WHERE order_id = $1;
	`
	rows, err := r.db.Query(ctx, itemsQuery, o.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}

	return &o, nil
}

// func (r *PgOrdersRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
//
// 	const query = `
// 		SELECT
// 			o.id, o.order_uid, o.track_number, o.entry,
// 			d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
// 			p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
// 			p.bank, p.delivery_cost, p.goods_total, p.custom_fee
// 		FROM orders o
// 		JOIN deliveries d ON o.id = d.order_id
// 		JOIN payments p ON o.id = p.order_id
// 		WHERE o.id = $1;
// 	`
//
// 	var o models.Order
// 	err := r.db.QueryRow(ctx, query, id).Scan(
// 		&o.ID, &o.OrderUID, &o.TrackNumber, &o.Entry,
// 		&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City,
// 		&o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email,
// 		&o.Payment.Transaction, &o.Payment.RequestID, &o.Payment.Currency, &o.Payment.Provider,
// 		&o.Payment.Amount, &o.Payment.PaymentDT, &o.Payment.Bank, &o.Payment.DeliveryCost,
// 		&o.Payment.GoodsTotal, &o.Payment.CustomFee,
// 	)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return nil, models.ErrOrderNotFound
// 		}
// 		return nil, err
// 	}
//
// 	// Подтягиваем items
// 	const itemsQuery = `
// 		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
// 		FROM items
// 		WHERE order_id = $1;
// 	`
// 	rows, err := r.db.Query(ctx, itemsQuery, id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
//
// 	for rows.Next() {
// 		var item models.Item
// 		if err := rows.Scan(
// 			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
// 			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
// 		); err != nil {
// 			return nil, err
// 		}
// 		o.Items = append(o.Items, item)
// 	}
//
// 	return &o, nil
// }
