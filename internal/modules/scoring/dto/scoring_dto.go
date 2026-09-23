package dto

import "time"

type ExamResultRecapDTO struct {
	ScheduleID      int        `json:"schedule_id"`
	ParticipantCode string     `json:"participant_code"`
	Name            string     `json:"name"`
	DeskNumber      int        `json:"desk_number"`
	LoginTime       *time.Time `json:"login_time,omitempty"`
	FinishTime      *time.Time `json:"finish_time,omitempty"`
	ExamStatus      string     `json:"exam_status"`
	FinalScore      float64    `json:"final_score"`
	PassingStatus   string     `json:"passing_status"`
	Ranking         int        `json:"ranking"`
}

type DashboardStatsDTO struct {
	TotalExamsHeld        int     `json:"total_exams_held"`
	PassingRatePercentage float64 `json:"passing_rate_percentage"`
	AverageScore          float64 `json:"average_score"`
	ZeroScorePercentage   float64 `json:"zero_score_percentage"`
}
