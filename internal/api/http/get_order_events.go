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

	if order != nil && order.IsFinal {
		c.JSON(http.StatusOK, order.Events)
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

	c.Stream(func(w io.Writer) bool {
		select {
		case event := <-clientChan:
			data, _ := json.Marshal(event)
			c.SSEvent("message", string(data))
			return true
		case <-client.Disconnect:
			return false
		}
	})
}

func newGetOrderEventsHandler(orders usecase.Orders, notifier domain.OrderObserver, timeout time.Duration) getOrderEventsHandler {
	return getOrderEventsHandler{orders: orders, notifier: notifier, timeout: timeout}
}
