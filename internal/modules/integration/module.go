package integration

import (
	"poltekkes-cat-backend/internal/modules/integration/handler"
	integrationMiddleware "poltekkes-cat-backend/internal/modules/integration/middleware"
	"poltekkes-cat-backend/internal/modules/integration/repository"
	"poltekkes-cat-backend/internal/modules/integration/service"
	"poltekkes-cat-backend/internal/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Init(apiV1 *gin.RouterGroup, db *sqlx.DB) {
	repo := repository.NewIntegrationRepository(db)
	svc := service.NewIntegrationService(repo)
	hdl := handler.NewIntegrationHandler(svc)

	// 1. Admin API Key Management Routes
	adminKeyGroup := apiV1.Group("/api-keys")
	adminKeyGroup.Use(middleware.AdminAuthMiddleware("admin", "cat"))
	{
		adminKeyGroup.GET("", hdl.GetAPIKeys)
		adminKeyGroup.POST("", hdl.CreateAPIKey)
		adminKeyGroup.PUT("/:id", hdl.UpdateAPIKey)
		adminKeyGroup.DELETE("/:id", hdl.DeleteAPIKey)
		adminKeyGroup.POST("/:id/regenerate", hdl.RegenerateAPIKey)
		adminKeyGroup.GET("/logs", hdl.GetAccessLogs)
	}

	// 2. SPMB M2M Integration Routes (Protected by API Key & Optional IP Whitelist)
	spmbGroup := apiV1.Group("/integration/spmb")
	spmbGroup.Use(integrationMiddleware.APIKeyAuthMiddleware(repo))
	{
		spmbGroup.GET("/exams", hdl.GetSPMBActiveExams)
		spmbGroup.POST("/register", hdl.RegisterSPMBParticipant)
	}
}
