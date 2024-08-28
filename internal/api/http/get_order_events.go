package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/therealyo/justdone/domain"
	"github.com/therealyo/justdone/internal/usecase"
)

type getOrderEventsRequest struct {
	OrderID string `uri:"order_id" form:"order_id" binding:"required,uuid"`
}

type getOrderEventsHandler struct {
	orders   usecase.Orders
	timeout  time.Duration
	notifier domain.OrderObserver
}

// GetOrderEventsHandler godoc
// @Summary      Stream order events
// @Description  Stream events for an order using Server-Side Events (SSE).
// @Tags         orders
// @Accept       json
// @Produce      text/event-stream
// @Param        order_id  path   string  true  "ID of the order"
// @Success      200  {object}  domain.OrderEvent  "Stream of order events"
// @Router       /orders/{order_id}/events [get]
func (h getOrderEventsHandler) handle(c *gin.Context) {
	var req getOrderEventsRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orders.GetOrder(req.OrderID)
	if err != nil && err != domain.ErrOrderNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Make buffered channel to avoid incorrect order of events
	clientChan := make(chan domain.OrderEvent, 1)
	client := domain.OrderEventsSubscriber{
		EventChan:  clientChan,
		Disconnect: make(chan bool),
		Timeout:    h.timeout,
	}

	h.notifier.RegisterClient(req.OrderID, client)
	defer h.notifier.UnregisterClient(req.OrderID, client)

	if order != nil {
		for _, event := range order.Events {
			if event.OrderStatus.Value() <= order.Status.Value() {
				data, _ := json.Marshal(event)
				h.notifier.AddProcessedEvent(req.OrderID, event)
				c.SSEvent("message", string(data))
				c.Writer.Flush()
			}
		}

		if order.IsFinal {
			fmt.Println("Order is final, closing connection")
			return
		}
	}

	fmt.Println("Starting stream")

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			fmt.Println("Client disconnected: close connection")
			return false
		case event, ok := <-clientChan:
			if !ok {
				fmt.Println("Client channel closed, stopping stream")
				return false
			}
			data, _ := json.Marshal(event)
			c.SSEvent("message", string(data))
			c.Writer.Flush()
			if event.IsFinal {
				fmt.Println("Order is final, closing connection")
				return false
			}
			return true

		case <-client.Disconnect:
			fmt.Println("Client disconnected")
			return false
		}
	})
}

func newGetOrderEventsHandler(orders usecase.Orders, notifier domain.OrderObserver, timeout time.Duration) getOrderEventsHandler {
	return getOrderEventsHandler{orders: orders, notifier: notifier, timeout: timeout}
}
