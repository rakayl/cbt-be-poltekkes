package examination

import (
	"poltekkes-cat-backend/internal/modules/examination/handler"
	"poltekkes-cat-backend/internal/modules/examination/repository"
	"poltekkes-cat-backend/internal/modules/examination/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Init(rg *gin.RouterGroup, db *sqlx.DB, siakadDBs ...*sqlx.DB) {
	repo := repository.NewExaminationRepository(db, siakadDBs...)
	svc := service.NewExaminationService(repo)
	hdl := handler.NewExaminationHandler(svc)
	hdl.RegisterRoutes(rg)
}
