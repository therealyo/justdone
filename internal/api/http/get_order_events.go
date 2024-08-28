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

	clientChan := make(chan domain.OrderEvent)
	client := domain.OrderEventsSubscriber{
		EventChan:  clientChan,
		Disconnect: make(chan bool),
		Timeout:    h.timeout,
	}

	h.notifier.RegisterClient(req.OrderID, client)
	defer h.notifier.UnregisterClient(req.OrderID, client)

	// Stream all current events regardless of order status
	if order != nil {
		fmt.Println("Order found", order.OrderID)
		for _, event := range order.Events {
			data, _ := json.Marshal(event)
			fmt.Println("Sending event to client", string(data))
			c.SSEvent("message", string(data))
			c.Writer.Flush() // Ensure data is sent immediately
		}

		// If the order is in a final state, close the connection
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
			fmt.Println("Sending event to client", string(data))
			c.SSEvent("message", string(data))
			c.Writer.Flush() // Ensure data is sent immediately
			return true

		case <-client.Disconnect:
			fmt.Println("Client timeout")
			return false
		}
	})
}

func newGetOrderEventsHandler(orders usecase.Orders, notifier domain.OrderObserver, timeout time.Duration) getOrderEventsHandler {
	return getOrderEventsHandler{orders: orders, notifier: notifier, timeout: timeout}
}
