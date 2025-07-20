package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func NewRouter(logger *zap.Logger) *gin.Engine {
	r := gin.New()

	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(params gin.LogFormatterParams) string {
			logger.Info(params.Request.URL.Path,
				zap.String("method", params.Method),
				zap.String("path", params.Path),
				zap.Int("status", params.StatusCode),
				zap.Duration("latency", params.Latency),
			)
			return ""
		},
	}))

	r.Use(gin.Recovery())

	url := ginSwagger.URL("/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	return r
}
