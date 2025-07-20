package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/skinkvi/effective_mobile/internal/handlers"
	"github.com/skinkvi/effective_mobile/internal/storage/postgres"
	"go.uber.org/zap"
)

func SubscriptionRoutes(r *gin.RouterGroup, storage *postgres.Storage, logger *zap.Logger) {
	handler := handlers.New(storage, logger)

	subscriptions := r.Group("/subscriptions")
	{
		subscriptions.POST("", handler.CreateSubscription)
		subscriptions.GET("/:id", handler.GetSubscription)
		subscriptions.PUT("/:id", handler.UpdateSubscription)
		subscriptions.DELETE("/:id", handler.DeleteSubscription)
		subscriptions.GET("", handler.ListSubscriptions)
		subscriptions.GET("/total_cost", handler.CalculateTotalCost)
	}
}
