package bank_soal

import (
	"poltekkes-cat-backend/internal/modules/bank_soal/handler"
	"poltekkes-cat-backend/internal/modules/bank_soal/repository"
	"poltekkes-cat-backend/internal/modules/bank_soal/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Init(rg *gin.RouterGroup, db *sqlx.DB) {
	// 1. Standard Static Bank Soal (Legacy Compatible)
	repo := repository.NewBankSoalRepository(db)
	svc := service.NewBankSoalService(repo)
	hdl := handler.NewBankSoalHandler(svc)
	hdl.RegisterRoutes(rg)

	// 2. Dynamic Bank Soal (Vignettes, Multimedia & Blueprints)
	dynRepo := repository.NewDynamicBankSoalRepository(db)
	dynSvc := service.NewDynamicBankSoalService(dynRepo)
	dynHdl := handler.NewDynamicBankSoalHandler(dynSvc)
	dynHdl.RegisterRoutes(rg.Group("/dynamic"))
}
