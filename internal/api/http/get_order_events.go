package http

import (
	"encoding/json"
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
	events  usecase.Events
	orders  usecase.Orders
	timeout time.Duration
}

// GetEventsHandler godoc
// @Summary Get order events
// @Description Get events for a specific order
// @Tags orders
// @Accept json
// @Produce json
// @Param order_id path string true "Order ID" format(uuid)
// @Success 200 {object} []domain.OrderEvent
// @Router /orders/{order_id}/events [get]
func (h getOrderEventsHandler) handle(c *gin.Context) {
	var req getOrderEventsRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	clientChan := make(chan domain.OrderEvent)
	// h.events.RegisterSSEClient(req.OrderID, clientChan)
	// defer h.events.UnregisterSSEClient(req.OrderID, clientChan)

	c.Stream(func(w io.Writer) bool {
		select {
		case event := <-clientChan:
			data, _ := json.Marshal(event)
			c.SSEvent("message", string(data))
			return true
		case <-time.After(h.timeout):
			c.SSEvent("error", "timeout")
			return false
		}
	})
}

func newGetOrderEventsHandler(events usecase.Events, orders usecase.Orders, timeout time.Duration) getOrderEventsHandler {
	return getOrderEventsHandler{events: events, orders: orders, timeout: timeout}
}
