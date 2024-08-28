package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/therealyo/justdone/domain"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) EventRepository {
	return EventRepository{db: db}
}

func (r EventRepository) Get(eventID string) (*domain.OrderEvent, error) {
	query := `SELECT event_id, order_id, user_id, order_status, created_at, updated_at, is_final 
			  FROM order_events WHERE event_id = $1`

	var event domain.OrderEvent

	err := r.db.QueryRow(query, eventID).Scan(
		&event.EventID,
		&event.OrderID,
		&event.UserID,
		&event.OrderStatus,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.IsFinal,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &event, nil
}

func (r EventRepository) Create(event domain.OrderEvent) error {
	query := `INSERT INTO order_events (event_id, order_id, user_id, order_status, created_at, updated_at, is_final)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(query,
		event.EventID,
		event.OrderID,
		event.UserID,
		event.OrderStatus,
		event.CreatedAt,
		event.UpdatedAt,
		event.IsFinal,
	)

	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

func (r EventRepository) Update(event domain.OrderEvent) error {
	fmt.Println("updating event", event)
	query := `UPDATE order_events SET order_status = $1, is_final = $2, updated_at = $3 WHERE event_id = $4`

	_, err := r.db.Exec(query,
		event.OrderStatus,
		event.IsFinal,
		event.UpdatedAt,
		event.EventID,
	)

	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}
	return nil
}

func (r EventRepository) Delete(eventID string) error {
	query := `DELETE FROM order_events WHERE event_id = $1`

	_, err := r.db.Exec(query, eventID)

	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	return nil
}

var _ domain.EventRepository = new(EventRepository)
