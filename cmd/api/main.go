package main

import (
	"log"
	"os"

	"poltekkes-cat-backend/internal/config"
	"poltekkes-cat-backend/internal/shared/database"
	"poltekkes-cat-backend/internal/shared/middleware"
	"poltekkes-cat-backend/internal/shared/response"

	// Feature-Driven Modular Packages
	"poltekkes-cat-backend/internal/modules/auth"
	"poltekkes-cat-backend/internal/modules/bank_soal"
	"poltekkes-cat-backend/internal/modules/berita"
	"poltekkes-cat-backend/internal/modules/english_proficiency"
	"poltekkes-cat-backend/internal/modules/examination"
	"poltekkes-cat-backend/internal/modules/integration"
	"poltekkes-cat-backend/internal/modules/periode"
	"poltekkes-cat-backend/internal/modules/peserta"
	"poltekkes-cat-backend/internal/modules/referensi"
	"poltekkes-cat-backend/internal/modules/scoring"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Application Configuration
	cfg := config.LoadConfig()

	// 2. Initialize PostgreSQL Database Connection Pool (CBT Engine)
	cbtDB := database.NewPostgresConnection(cfg)
	defer cbtDB.Close()

	// 2b. Initialize SIAKAD Database Connection Pool (gate, ref, pendaftaran)
	siakadDB := database.NewSiakadConnection(cfg, cbtDB)
	defer func() {
		if siakadDB != cbtDB {
			_ = siakadDB.Close()
		}
	}()

	// 3. Setup Gin HTTP Engine
	r := gin.Default()

	// 4. Global Middlewares
	r.Use(middleware.TraceIDMiddleware())
	r.Use(middleware.CORSMiddleware())

	// 5. Static File Serving for Multimedia (Audio, Images, Video)
	_ = os.MkdirAll("./uploads/media", 0755)
	r.Static("/uploads", "./uploads")

	// 6. Healthcheck Endpoint
	r.GET("/api/v1/health", func(c *gin.Context) {
		dbMode := "Single Database (" + cfg.DBName + ")"
		if cfg.IsDualDB() {
			dbMode = "Dual Database (CBT: " + cfg.DBName + ", SIAKAD: " + cfg.DBSiakadName + ")"
		}
		response.Success(c, "POLTEKKES Enterprise Modular CBT API Engine is healthy", map[string]string{
			"version":       "2.0.0",
			"architecture":  "Feature-Driven Modular Monolith (Dual-Database Capable)",
			"database_mode": dbMode,
			"status":        "HEALTHY",
		})
	})

	// 7. Register Feature Modules on /api/v1
	apiV1 := r.Group("/api/v1")
	{
		// Module 1: Auth & Gate SSO
		auth.Init(apiV1.Group("/auth"), cbtDB, siakadDB)

		// Module 2: Master Periode
		periode.Init(apiV1.Group("/periods"), cbtDB)

		// Module 3: Bank Soal & Pertanyaan
		bank_soal.Init(apiV1.Group("/bank-soal"), cbtDB)

		// Module 4: Daftar Peserta
		peserta.Init(apiV1.Group("/peserta"), cbtDB, siakadDB)
		peserta.Init(apiV1.Group("/participants"), cbtDB, siakadDB)

		// Module 5: Examination (Pelaksanaan & Pengerjaan Ujian)
		examination.Init(apiV1.Group("/exams"), cbtDB, siakadDB)

		// Module 6: Scoring, Rekapitulasi & Dashboard Stats
		scoring.Init(apiV1.Group("/scoring"), cbtDB)

		// Module 7: Berita & Pengumuman
		berita.Init(apiV1.Group("/berita"), cbtDB)
		berita.Init(apiV1.Group("/announcements"), cbtDB)

		// Module 8: Master Referensi (Ruang, Skor, Wilayah, Panitia, Jenis Ujian, Salin Soal)
		referensi.Init(apiV1.Group("/referensi"), cbtDB, siakadDB)
		referensi.Init(apiV1.Group("/reference"), cbtDB, siakadDB)

		// Module 9: TOEFL & TOEIC English Proficiency Testing Module
		english_proficiency.Init(apiV1.Group("/ep"), cbtDB)

		// Module 10: Integrasi Eksternal & M2M API Key (SPMB)
		integration.Init(apiV1, cbtDB)

		// Backwards-Compatible CAT Admin & CAT Front Aliases for existing UI paths
		catAdminGroup := apiV1.Group("/cat")
		{
			scoring.Init(catAdminGroup, cbtDB)
			periode.Init(catAdminGroup.Group("/periods"), cbtDB)
			bank_soal.Init(catAdminGroup.Group("/question-banks"), cbtDB)
			examination.Init(catAdminGroup.Group("/exams"), cbtDB, siakadDB)
			referensi.Init(catAdminGroup.Group("/referensi"), cbtDB, siakadDB)
			english_proficiency.Init(catAdminGroup.Group("/ep"), cbtDB)
			peserta.Init(catAdminGroup.Group("/peserta"), cbtDB, siakadDB)
			peserta.Init(catAdminGroup.Group("/participants"), cbtDB, siakadDB)
			integration.Init(catAdminGroup, cbtDB)
		}

		catFrontGroup := apiV1.Group("/catfront")
		{
			auth.Init(catFrontGroup.Group("/auth"), cbtDB, siakadDB)
			examination.Init(catFrontGroup.Group("/exams"), cbtDB, siakadDB)
		}
	}

	log.Printf("🚀 POLTEKKES Enterprise Modular CBT API running on port :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
