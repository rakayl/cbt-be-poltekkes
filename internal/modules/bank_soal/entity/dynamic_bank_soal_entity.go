package entity

import (
	"encoding/json"
	"time"
)

// BankStimulus represents a clinical case narrative, reading passage, or multimedia stimulus unit
type BankStimulus struct {
	IDStimulus      int64           `db:"id_stimulus" json:"id_stimulus"`
	KodeSoal        string          `db:"kodesoal" json:"kodesoal"`
	StimulusCode    string          `db:"stimulus_code" json:"stimulus_code"`
	Title           string          `db:"title" json:"title"`
	NarrativeText   string          `db:"narrative_text" json:"narrative_text"`
	MediaType       string          `db:"media_type" json:"media_type"` // NONE, AUDIO, IMAGE, VIDEO, MULTI
	MediaURL        *string         `db:"media_url" json:"media_url"`
	MediaMetadata   json.RawMessage `db:"media_metadata" json:"media_metadata"`
	CategoryTag     string          `db:"category_tag" json:"category_tag"`
	DifficultyLevel int             `db:"difficulty_level" json:"difficulty_level"` // 1=Easy, 2=Medium, 3=Hard
	SortOrder       int             `db:"sort_order" json:"sort_order"`
	SoftDelete      string          `db:"softdelete" json:"softdelete"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`

	// Relational Children
	Items []*StimulusItem `json:"items,omitempty"`
}

// StimulusItem represents a sub-question under a stimulus
type StimulusItem struct {
	IDItem            int64     `db:"id_item" json:"id_item"`
	IDStimulus        int64     `db:"id_stimulus" json:"id_stimulus"`
	ItemOrder         int       `db:"item_order" json:"item_order"`
	QuestionText      string    `db:"question_text" json:"question_text"`
	QuestionMediaType string    `db:"question_media_type" json:"question_media_type"` // NONE, AUDIO, IMAGE, VIDEO
	QuestionMediaURL  *string   `db:"question_media_url" json:"question_media_url"`
	ItemType          string    `db:"item_type" json:"item_type"` // SINGLE_CHOICE, MULTI_CHOICE
	WeightCorrect     float64   `db:"weight_correct" json:"weight_correct"`
	WeightWrong       float64   `db:"weight_wrong" json:"weight_wrong"`
	WeightBlank       float64   `db:"weight_blank" json:"weight_blank"`
	CorrectAnswer     string    `db:"correct_answer" json:"correct_answer"` // e.g. "A" or "1,3"
	Explanation       *string   `db:"explanation" json:"explanation"`
	SoftDelete        string    `db:"softdelete" json:"softdelete"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`

	// Relational Options
	Options []*ItemOption `json:"options,omitempty"`
}

// ItemOption represents an option choice A-E under a sub-question
type ItemOption struct {
	IDOption    int64     `db:"id_option" json:"id_option"`
	IDItem      int64     `db:"id_item" json:"id_item"`
	OptionLabel string    `db:"option_label" json:"option_label"` // A, B, C, D, E
	OptionText  string    `db:"option_text" json:"option_text"`
	MediaType   string    `db:"media_type" json:"media_type"` // NONE, AUDIO, IMAGE
	MediaURL    *string   `db:"media_url" json:"media_url"`
	SortOrder   int       `db:"sort_order" json:"sort_order"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// ExamBlueprint represents a blueprint for dynamic exam composition
type ExamBlueprint struct {
	IDBlueprint          int64            `db:"id_blueprint" json:"id_blueprint"`
	BlueprintCode        string           `db:"blueprint_code" json:"blueprint_code"`
	Title                string           `db:"title" json:"title"`
	Description          *string          `db:"description" json:"description"`
	TotalTargetQuestions int              `db:"total_target_questions" json:"total_target_questions"`
	PassingScore         float64          `db:"passing_score" json:"passing_score"`
	DurationMinutes      int              `db:"duration_minutes" json:"duration_minutes"`
	SoftDelete           string           `db:"softdelete" json:"softdelete"`
	CreatedAt            time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time        `db:"updated_at" json:"updated_at"`
	Rules                []*BlueprintRule `json:"rules,omitempty"`
}

// BlueprintRule represents a rule specifying quotas per tag and difficulty
type BlueprintRule struct {
	IDRule           int64     `db:"id_rule" json:"id_rule"`
	IDBlueprint      int64     `db:"id_blueprint" json:"id_blueprint"`
	CategoryTag      string    `db:"category_tag" json:"category_tag"`
	DifficultyLevel  int       `db:"difficulty_level" json:"difficulty_level"` // 0=Any, 1=Easy, 2=Medium, 3=Hard
	QuotaCount       int       `db:"quota_count" json:"quota_count"`
	WeightMultiplier float64   `db:"weight_multiplier" json:"weight_multiplier"`
	SortOrder        int       `db:"sort_order" json:"sort_order"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

// ExamScheduleExt represents the extended time window and scoring configurations
type ExamScheduleExt struct {
	IDJadwalUjian         int       `db:"idjadwalujian" json:"idjadwalujian"`
	IDBlueprint           *int64    `db:"id_blueprint" json:"id_blueprint"`
	WindowStartTime       time.Time `db:"window_start_time" json:"window_start_time"`
	WindowEndTime         time.Time `db:"window_end_time" json:"window_end_time"`
	ScoringRule           string    `db:"scoring_rule" json:"scoring_rule"` // STANDARD, PENALTY_MINUS_ONE, CUSTOM
	DefaultCorrectScore   float64   `db:"default_correct_score" json:"default_correct_score"`
	DefaultWrongScore     float64   `db:"default_wrong_score" json:"default_wrong_score"`
	DefaultBlankScore     float64   `db:"default_blank_score" json:"default_blank_score"`
	AutoSubmitOnWindowEnd bool      `db:"auto_submit_on_window_end" json:"auto_submit_on_window_end"`
	Timezone              string    `db:"timezone" json:"timezone"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time `db:"updated_at" json:"updated_at"`
}

// ParticipantSessionExt represents the live participant exam session
type ParticipantSessionExt struct {
	IDSession          int64           `db:"id_session" json:"id_session"`
	IDJadwalUjian      int             `db:"idjadwalujian" json:"idjadwalujian"`
	KodePeserta        string          `db:"kodepeserta" json:"kodepeserta"`
	SessionStatus      string          `db:"session_status" json:"session_status"` // NOT_STARTED, IN_PROGRESS, SUBMITTED, EXPIRED_AUTO_SUBMIT, LOCKED
	StartedAt          *time.Time      `db:"started_at" json:"started_at"`
	CompletedAt        *time.Time      `db:"completed_at" json:"completed_at"`
	CalculatedDeadline *time.Time      `db:"calculated_deadline" json:"calculated_deadline"`
	SessionSnapshot    json.RawMessage `db:"session_snapshot" json:"session_snapshot"`
	TotalScore         float64         `db:"total_score" json:"total_score"`
	TotalCorrect       int             `db:"total_correct" json:"total_correct"`
	TotalWrong         int             `db:"total_wrong" json:"total_wrong"`
	TotalUnanswered    int             `db:"total_unanswered" json:"total_unanswered"`
	PassStatus         string          `db:"pass_status" json:"pass_status"`
	SubmitReason       *string         `db:"submit_reason" json:"submit_reason"`
	ClientIP           *string         `db:"client_ip" json:"client_ip"`
	UserAgent          *string         `db:"user_agent" json:"user_agent"`
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time       `db:"updated_at" json:"updated_at"`
}
