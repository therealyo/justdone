package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/therealyo/justdone/domain"
	"github.com/therealyo/justdone/internal/usecase"
)

type postEventHandler struct {
	events usecase.Events
}

type postEventRequest struct {
	EventID     string    `json:"event_id" binding:"required,uuid"`
	OrderID     string    `json:"order_id" binding:"required,uuid"`
	UserID      string    `json:"user_id" binding:"required,uuid"`
	OrderStatus string    `json:"order_status" binding:"required,oneof=cool_order_created sbu_verification_pending confirmed_by_mayor changed_my_mind failed chinazes give_my_money_back"`
	CreatedAt   time.Time `json:"created_at" binding:"required"`
	UpdatedAt   time.Time `json:"updated_at" binding:"required"`
}

// PostEventHandler godoc
// @Summary      handle event from JustPay!
// @Description  handle event from JustPay!
// @Tags         webhooks
// @Accept       json
// @Produce      json
// @Param        event   body      postEventRequest  true  "Event"
// @Success      200  {object}  domain.OrderEvent
// @Router       /webhooks/payments/orders [post]
func (h postEventHandler) handle(c *gin.Context) {
	var req postEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.events.Create(&domain.OrderEvent{
		EventID:     req.EventID,
		OrderID:     req.OrderID,
		UserID:      req.UserID,
		OrderStatus: domain.OrderStatus(req.OrderStatus),
		CreatedAt:   req.CreatedAt,
		UpdatedAt:   req.UpdatedAt,
		IsFinal:     domain.OrderStatus(req.OrderStatus).IsFinal(),
	})

	switch {
	case err == domain.ErrEventConflict:
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
	case err == domain.ErrOrderAlreadyFinal:
		c.AbortWithStatusJSON(http.StatusGone, gin.H{"error": err.Error()})
	case err == domain.ErrOrderNotFound:
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case err != nil:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusOK, gin.H{"message": "Event created"})
	}
}

func newPostEventHandler(events usecase.Events) postEventHandler {
	return postEventHandler{events: events}
}
