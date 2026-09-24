package dto

type CreatePeriodRequestDTO struct {
	PeriodName string `json:"period_name" binding:"required"`
	PeriodType string `json:"period_type"`
	IsOnline   int    `json:"is_online"`
}

type UpdatePeriodRequestDTO struct {
	PeriodName string `json:"period_name" binding:"required"`
	PeriodType string `json:"period_type"`
	IsOnline   int    `json:"is_online"`
}

type PeriodResponseDTO struct {
	PeriodID    int    `json:"period_id"`
	IDPeriode   int    `json:"idperiode"`
	PeriodName  string `json:"period_name"`
	NamaPeriode string `json:"namaperiode"`
	PeriodType  string `json:"period_type"`
	IsOnline    int    `json:"is_online"`
}
