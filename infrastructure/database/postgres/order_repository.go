package postgres

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/therealyo/justdone/domain"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return OrderRepository{db: db}
}

func (r OrderRepository) GetMany(filter *domain.OrderFilter) ([]domain.Order, error) {
	panic("unimplemented")
}

func (r OrderRepository) Get(orderID string) (*domain.Order, error) {
	query := `
		SELECT o.order_id, o.user_id, o.status, o.is_final, o.created_at, o.updated_at,
		       e.event_id, e.user_id, e.order_status, e.created_at, e.updated_at, e.is_final
		FROM orders o
		LEFT JOIN order_events e ON o.order_id = e.order_id
		WHERE o.order_id = $1
		ORDER BY e.created_at ASC
	`

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Error closing rows: %v\n", err)
		}
	}()

	var order *domain.Order

	for rows.Next() {
		var (
			eventID     sql.NullString
			orderStatus sql.NullString
			eventTime   sql.NullTime
			updateTime  sql.NullTime
			isFinal     sql.NullBool
			userID      sql.NullString
		)

		if order == nil {
			order = &domain.Order{}
			if err := rows.Scan(
				&order.OrderID,
				&order.UserID,
				&order.Status,
				&order.IsFinal,
				&order.CreatedAt,
				&order.UpdatedAt,
				&eventID,
				&userID,
				&orderStatus,
				&eventTime,
				&updateTime,
				&isFinal,
			); err != nil {
				return nil, fmt.Errorf("failed to scan order row: %w", err)
			}
		} else {
			if err := rows.Scan(
				new(string),
				new(string),
				new(string),
				new(bool),
				new(time.Time),
				new(time.Time),
				&eventID,
				&userID,
				&orderStatus,
				&eventTime,
				&updateTime,
				&isFinal,
			); err != nil {
				return nil, fmt.Errorf("failed to scan event row: %w", err)
			}
		}

		if eventID.Valid {
			event := domain.OrderEvent{
				EventID:     eventID.String,
				OrderID:     order.OrderID,
				UserID:      userID.String,
				OrderStatus: domain.OrderStatus(orderStatus.String),
				CreatedAt:   eventTime.Time,
				UpdatedAt:   updateTime.Time,
				IsFinal:     isFinal.Bool,
			}
			order.Events = append(order.Events, event)
		}
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", rows.Err())
	}

	if order == nil {
		return nil, nil
	}

	if len(order.Events) > 0 {
		order.LastEvent = &order.Events[len(order.Events)-1]
	}

	return order, nil
}

func (r OrderRepository) Save(order *domain.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("rollback error: %v, original error: %w", rbErr, err)
			}
		} else {
			err = tx.Commit()
		}
	}()

	orderQuery := `
		INSERT INTO orders (order_id, user_id, status, is_final, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (order_id) DO UPDATE 
		SET user_id = EXCLUDED.user_id,
			status = EXCLUDED.status,
			is_final = EXCLUDED.is_final,
			created_at = EXCLUDED.created_at,
			updated_at = EXCLUDED.updated_at
	`

	_, err = tx.Exec(orderQuery,
		order.OrderID,
		order.UserID,
		order.Status,
		order.IsFinal,
		order.CreatedAt,
		order.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	return nil
}

var _ domain.OrderRepository = new(OrderRepository)
