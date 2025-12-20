package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/dunooo0ooo/wb-tech-l0/internal/infrastructure"
	"time"

	"github.com/dunooo0ooo/wb-tech-l0/internal/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, o *entity.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard, updated_at
		) VALUES (
			@order_uid, @track_number, @entry, @locale, @internal_signature,
			@customer_id, @delivery_service, @shardkey, @sm_id, @date_created, @oof_shard, now()
		)
		ON CONFLICT (order_uid) DO UPDATE SET
			track_number=EXCLUDED.track_number,
			entry=EXCLUDED.entry,
			locale=EXCLUDED.locale,
			internal_signature=EXCLUDED.internal_signature,
			customer_id=EXCLUDED.customer_id,
			delivery_service=EXCLUDED.delivery_service,
			shardkey=EXCLUDED.shardkey,
			sm_id=EXCLUDED.sm_id,
			date_created=EXCLUDED.date_created,
			oof_shard=EXCLUDED.oof_shard,
			updated_at=now()
	`, pgx.NamedArgs{
		"order_uid":          o.OrderUID,
		"track_number":       o.TrackNumber,
		"entry":              o.Entry,
		"locale":             o.Locale,
		"internal_signature": o.InternalSignature,
		"customer_id":        o.CustomerID,
		"delivery_service":   o.DeliveryService,
		"shardkey":           o.ShardKey,
		"sm_id":              o.SmID,
		"date_created":       o.DateCreated,
		"oof_shard":          o.OofShard,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
		VALUES (@order_uid,@name,@phone,@zip,@city,@address,@region,@email)
		ON CONFLICT (order_uid) DO UPDATE SET
			name=EXCLUDED.name,
			phone=EXCLUDED.phone,
			zip=EXCLUDED.zip,
			city=EXCLUDED.city,
			address=EXCLUDED.address,
			region=EXCLUDED.region,
			email=EXCLUDED.email
	`, pgx.NamedArgs{
		"order_uid": o.OrderUID,
		"name":      o.Delivery.Name,
		"phone":     o.Delivery.Phone,
		"zip":       o.Delivery.Zip,
		"city":      o.Delivery.City,
		"address":   o.Delivery.Address,
		"region":    o.Delivery.Region,
		"email":     o.Delivery.Email,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO payment (
			order_uid, transaction, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES (
			@order_uid, @transaction, @request_id, @currency, @provider,
			@amount, @payment_dt, @bank, @delivery_cost, @goods_total, @custom_fee
		)
		ON CONFLICT (order_uid) DO UPDATE SET
			transaction=EXCLUDED.transaction,
			request_id=EXCLUDED.request_id,
			currency=EXCLUDED.currency,
			provider=EXCLUDED.provider,
			amount=EXCLUDED.amount,
			payment_dt=EXCLUDED.payment_dt,
			bank=EXCLUDED.bank,
			delivery_cost=EXCLUDED.delivery_cost,
			goods_total=EXCLUDED.goods_total,
			custom_fee=EXCLUDED.custom_fee
	`, pgx.NamedArgs{
		"order_uid":     o.OrderUID,
		"transaction":   o.Payment.Transaction,
		"request_id":    o.Payment.RequestID,
		"currency":      o.Payment.Currency,
		"provider":      o.Payment.Provider,
		"amount":        o.Payment.Amount,
		"payment_dt":    o.Payment.PaymentDT,
		"bank":          o.Payment.Bank,
		"delivery_cost": o.Payment.DeliveryCost,
		"goods_total":   o.Payment.GoodsTotal,
		"custom_fee":    o.Payment.CustomFee,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
	}

	b := &pgx.Batch{}
	for _, it := range o.Items {
		if it.ChrtID == 0 {
			continue
		}

		b.Queue(`
			INSERT INTO items (
				order_uid, chrt_id, track_number, price, rid, name, sale, size,
				total_price, nm_id, brand, status
			) VALUES (
				@order_uid, @chrt_id, @track_number, @price, @rid, @name, @sale, @size,
				@total_price, @nm_id, @brand, @status
			)
			ON CONFLICT (order_uid, chrt_id) DO UPDATE SET
				track_number=EXCLUDED.track_number,
				price=EXCLUDED.price,
				rid=EXCLUDED.rid,
				name=EXCLUDED.name,
				sale=EXCLUDED.sale,
				size=EXCLUDED.size,
				total_price=EXCLUDED.total_price,
				nm_id=EXCLUDED.nm_id,
				brand=EXCLUDED.brand,
				status=EXCLUDED.status
		`, pgx.NamedArgs{
			"order_uid":    o.OrderUID,
			"chrt_id":      it.ChrtID,
			"track_number": it.TrackNumber,
			"price":        it.Price,
			"rid":          it.Rid,
			"name":         it.Name,
			"sale":         it.Sale,
			"size":         it.Size,
			"total_price":  it.TotalPrice,
			"nm_id":        it.NmID,
			"brand":        it.Brand,
			"status":       it.Status,
		})
	}

	br := tx.SendBatch(ctx, b)
	if err := br.Close(); err != nil {
		return fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	const q = `
		SELECT
			o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
			o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
			d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
			p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
			p.bank, p.delivery_cost, p.goods_total, p.custom_fee
		FROM orders o
		LEFT JOIN delivery d ON d.order_uid = o.order_uid
		LEFT JOIN payment  p ON p.order_uid = o.order_uid
		WHERE o.order_uid = @id
	`

	var o entity.Order
	var created time.Time

	err := r.pool.QueryRow(ctx, q, pgx.NamedArgs{"id": id}).Scan(
		&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature,
		&o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &created, &o.OofShard,
		&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City, &o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email,
		&o.Payment.Transaction, &o.Payment.RequestID, &o.Payment.Currency, &o.Payment.Provider, &o.Payment.Amount, &o.Payment.PaymentDT,
		&o.Payment.Bank, &o.Payment.DeliveryCost, &o.Payment.GoodsTotal, &o.Payment.CustomFee,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructure.ErrOrderNotFound
		}
		return nil, fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
	}
	o.DateCreated = created

	rows, err := r.pool.Query(ctx, `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items
		WHERE order_uid = @id
		ORDER BY chrt_id
	`, pgx.NamedArgs{"id": id})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
	}
	defer rows.Close()

	items := make([]entity.Item, 0, 8)
	for rows.Next() {
		var it entity.Item
		if err := rows.Scan(
			&it.ChrtID, &it.TrackNumber, &it.Price, &it.Rid, &it.Name,
			&it.Sale, &it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status,
		); err != nil {
			return nil, fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
	}
	o.Items = items

	return &o, nil
}

func (r *Repository) LoadRecent(ctx context.Context, limit int) ([]entity.Order, error) {
	if limit <= 0 {
		limit = 1000
	}

	rows, err := r.pool.Query(ctx, `
		SELECT order_uid
		FROM orders
		ORDER BY updated_at DESC
		LIMIT @lim
	`, pgx.NamedArgs{"lim": limit})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
	}
	defer rows.Close()

	ids := make([]string, 0, limit)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", infrastructure.ErrInternalDatabase, err)
	}

	out := make([]entity.Order, 0, len(ids))
	for _, id := range ids {
		o, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, *o)
	}
	return out, nil
}
