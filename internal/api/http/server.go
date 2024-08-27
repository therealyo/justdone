package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/therealyo/justdone/docs"
	"github.com/therealyo/justdone/internal/app"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/therealyo/justdone/docs"
)

type Server struct {
	app    *app.Application
	router *gin.Engine
}

func NewServer(app *app.Application) *Server {
	return &Server{
		app:    app,
		router: gin.Default(),
	}
}

func (s *Server) Setup() (*Server, error) {
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}

	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "OK",
		})
	})

	s.router.POST(
		"/webhooks/payments/orders",
		newPostEventHandler(s.app.Events).handle,
	)

	ordersGroup := s.router.Group("/orders")

	ordersGroup.GET(
		":order_id/events",
		newGetOrderEventsHandler(s.app.Orders, s.app.Notifier, 1*time.Minute).handle,
	)
	ordersGroup.GET(
		"",
		newGetOrdersHandler(s.app.Orders).handle,
	)

	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return s, nil
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
