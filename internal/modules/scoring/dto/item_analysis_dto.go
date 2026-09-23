package dto

import "time"

// DistractorOptionDTO represents the choice distribution for a single option
type DistractorOptionDTO struct {
	OptionKey       int     `json:"option_key"`
	OptionLabel     string  `json:"option_label"` // A, B, C, D, E
	OptionText      string  `json:"option_text"`
	IsCorrect       bool    `json:"is_correct"`
	TotalChosen     int     `json:"total_chosen"`
	ChosenPercent   float64 `json:"chosen_percent"`
	UpperGroupCount int     `json:"upper_group_count"` // 27% upper scorers
	LowerGroupCount int     `json:"lower_group_count"` // 27% lower scorers
	IsFunctioning   bool    `json:"is_functioning"`    // true if chosen by >= 5% in lower group (if distractor)
}

// ItemAnalysisItemDTO represents psychometrics for one question
type ItemAnalysisItemDTO struct {
	QuestionNumber      int                   `json:"question_number"`
	QuestionText        string                `json:"question_text"`
	CategoryTag         string                `json:"category_tag"`
	CorrectOptionKey    int                   `json:"correct_option_key"`
	TotalTakers         int                   `json:"total_takers"`
	CorrectCount        int                   `json:"correct_count"`
	WrongCount          int                   `json:"wrong_count"`
	DifficultyIndex     float64               `json:"difficulty_index"`     // P = Total Correct / Total Takers (0.00 - 1.00)
	DifficultyCategory  string                `json:"difficulty_category"`  // MUDAH (>0.70), SEDANG (0.30-0.70), SUKAR (<0.30)
	DiscriminationIndex float64               `json:"discrimination_index"` // D = (Upper Correct - Lower Correct) / (0.27 * Total)
	DiscriminationClass string                `json:"discrimination_class"` // SANGAT BAIK (>=0.40), BAIK (0.30-0.39), CUKUP/REVISI (0.20-0.29), BURUK (<0.20)
	Recommendation      string                `json:"recommendation"`       // DAPAT DIGUNAKAN, PERLU REVISI, BUANG / DIGANTI
	Distractors         []DistractorOptionDTO `json:"distractors"`
}

// ItemAnalysisSummaryDTO represents aggregate metrics of the exam
type ItemAnalysisSummaryDTO struct {
	ScheduleID         int     `json:"schedule_id"`
	ExamName           string  `json:"exam_name"`
	QuestionBankCode   string  `json:"question_bank_code"`
	TotalQuestions     int     `json:"total_questions"`
	TotalParticipants  int     `json:"total_participants"`
	EasyCount          int     `json:"easy_count"`
	MediumCount        int     `json:"medium_count"`
	HardCount          int     `json:"hard_count"`
	ExcellentDiscCount int     `json:"excellent_disc_count"`
	GoodDiscCount      int     `json:"good_disc_count"`
	FairDiscCount      int     `json:"fair_disc_count"`
	PoorDiscCount      int     `json:"poor_disc_count"`
	AverageDifficulty  float64 `json:"average_difficulty"`
	ExamReliabilityEst float64 `json:"exam_reliability_est"` // KR-20 estimate
}

// ItemAnalysisResponseDTO is the complete response for item analysis
type ItemAnalysisResponseDTO struct {
	Summary   ItemAnalysisSummaryDTO `json:"summary"`
	Questions []ItemAnalysisItemDTO  `json:"questions"`
}

// BeritaAcaraParticipantRow represents one participant row in Berita Acara
type BeritaAcaraParticipantRow struct {
	ParticipantCode string     `json:"participant_code"`
	Name            string     `json:"name"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	Score           float64    `json:"score"`
	Status          string     `json:"status"` // LULUS / TIDAK LULUS / TIDAK HADIR
	ViolationCount  int        `json:"violation_count"`
}

// BeritaAcaraResponseDTO represents official examination report data
type BeritaAcaraResponseDTO struct {
	ScheduleID        int                         `json:"schedule_id"`
	ExamName          string                      `json:"exam_name"`
	PeriodName        string                      `json:"period_name"`
	RoomName          string                      `json:"room_name"`
	ExamDate          string                      `json:"exam_date"`
	StartTime         string                      `json:"start_time"`
	EndTime           string                      `json:"end_time"`
	PassingGrade      float64                     `json:"passing_grade"`
	TotalRegistered   int                         `json:"total_registered"`
	TotalAttended     int                         `json:"total_attended"`
	TotalAbsent       int                         `json:"total_absent"`
	TotalPassed       int                         `json:"total_passed"`
	TotalFailed       int                         `json:"total_failed"`
	PassPercentage    float64                     `json:"pass_percentage"`
	HighestScore      float64                     `json:"highest_score"`
	LowestScore       float64                     `json:"lowest_score"`
	AverageScore      float64                     `json:"average_score"`
	ProctorNotes      string                      `json:"proctor_notes"`
	Supervisors       []string                    `json:"supervisors"`
	Participants      []BeritaAcaraParticipantRow `json:"participants"`
}
