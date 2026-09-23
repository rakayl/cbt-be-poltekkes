package periode

import (
	"poltekkes-cat-backend/internal/modules/periode/handler"
	"poltekkes-cat-backend/internal/modules/periode/repository"
	"poltekkes-cat-backend/internal/modules/periode/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Init(rg *gin.RouterGroup, db *sqlx.DB) {
	repo := repository.NewPeriodeRepository(db)
	svc := service.NewPeriodeService(repo)
	hdl := handler.NewPeriodeHandler(svc)
	hdl.RegisterRoutes(rg)
}
