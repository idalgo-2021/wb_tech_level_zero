package repository

import (
	"context"
	"errors"
	"wb_tech_level_zero/internal/orders"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrdersRepository struct {
	db *pgxpool.Pool
}

func NewOrdersRepository(db *pgxpool.Pool) *OrdersRepository {
	return &OrdersRepository{db: db}
}

func (r *OrdersRepository) GetOrderByUID(ctx context.Context, orderUID string) (*orders.Order, error) {
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

	var o orders.Order
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
			return nil, orders.ErrOrderNotFound
		}
		return nil, err
	}

	// items
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
		var item orders.Item
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

func (r *OrdersRepository) GetOrders(ctx context.Context, limit, offset int) ([]*orders.Order, int, error) {

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
			p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
			o.locale, o.internal_signature, o.customer_id, o.delivery_service,
			o.shardkey, o.sm_id, o.date_created, o.oof_shard
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

	var ordersList []*orders.Order

	for rows.Next() {
		var o orders.Order
		if err := rows.Scan(
			&o.ID, &o.OrderUID, &o.TrackNumber, &o.Entry,
			&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City,
			&o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email,
			&o.Payment.Transaction, &o.Payment.RequestID, &o.Payment.Currency, &o.Payment.Provider,
			&o.Payment.Amount, &o.Payment.PaymentDT, &o.Payment.Bank, &o.Payment.DeliveryCost,
			&o.Payment.GoodsTotal, &o.Payment.CustomFee,
			&o.Locale, &o.InternalSignature, &o.CustomerID, &o.DeliveryService,
			&o.Shardkey, &o.SmID, &o.DateCreated, &o.OofShard,
		); err != nil {
			return nil, 0, err
		}
		ordersList = append(ordersList, &o)
	}

	// without Items

	return ordersList, total, nil
}

func (r *OrdersRepository) SaveOrder(ctx context.Context, order *orders.Order) error {

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	var existingID int
	err = tx.QueryRow(ctx, `SELECT id FROM orders WHERE order_uid=$1`, order.OrderUID).Scan(&existingID)
	if err == nil {
		return orders.ErrOrderAlreadyExists
	} else if err != pgx.ErrNoRows {
		return err
	}

	var orderID int
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id
	`,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	).Scan(&orderID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO deliveries (
			order_id, name, phone, zip, city, address, region, email
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`,
		orderID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO payments (
			order_id, transaction, request_id, currency, provider, amount,
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`,
		orderID,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDT,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return err
	}

	for _, it := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO items (
				order_id, chrt_id, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`,
			orderID,
			it.ChrtID,
			it.TrackNumber,
			it.Price,
			it.Rid,
			it.Name,
			it.Sale,
			it.Size,
			it.TotalPrice,
			it.NmID,
			it.Brand,
			it.Status,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
