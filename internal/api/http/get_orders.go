package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/therealyo/justdone/domain"
	"github.com/therealyo/justdone/internal/usecase"
)

type getOrdersHandler struct {
	orders usecase.Orders
}

type getOrdersRequest struct {
	Status    []string `form:"status"`
	UserID    string   `form:"user_id"`
	Limit     int      `form:"limit,default=10"`
	Offset    int      `form:"offset,default=0"`
	IsFinal   *bool    `form:"is_final"`
	SortBy    string   `form:"sort_by,default=created_at"`
	SortOrder string   `form:"sort_order,default=desc"`
}

func (h getOrdersHandler) handle(c *gin.Context) {
	var req getOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filters, err := req.build()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orders, err := h.orders.GetOrders(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (req getOrdersRequest) validate() error {
	if len(req.Status) > 0 && req.IsFinal != nil {
		return fmt.Errorf("cannot specify both status and is_final")
	}

	if len(req.Status) == 0 && req.IsFinal == nil {
		return fmt.Errorf("must specify either status or is_final")
	}

	return nil
}

func (req getOrdersRequest) build() (*domain.OrderFilter, error) {
	filterOptions := []domain.FilterOption{
		domain.WithUserID(req.UserID),
		domain.WithLimit(req.Limit),
		domain.WithOffset(req.Offset),
		domain.WithSortBy(req.SortBy),
		domain.WithSortOrder(req.SortOrder),
	}

	if len(req.Status) > 0 {
		statuses := make([]domain.OrderStatus, len(req.Status))
		for i, s := range req.Status {
			status, err := domain.ParseOrderStatus(s)
			if err != nil {
				return nil, err
			}
			statuses[i] = status
		}
		filterOptions = append(filterOptions, domain.WithStatus(statuses...))
	}

	if req.IsFinal != nil {
		filterOptions = append(filterOptions, domain.WithIsFinal(*req.IsFinal))
	}

	return domain.NewOrderFilter(filterOptions...), nil
}

func newGetOrdersHandler(orders usecase.Orders) getOrdersHandler {
	return getOrdersHandler{orders: orders}
}
