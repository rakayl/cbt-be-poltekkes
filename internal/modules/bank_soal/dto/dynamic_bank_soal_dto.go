package dto

// OptionInputDTO represents an input option choice A-E
type OptionInputDTO struct {
	OptionLabel string  `json:"option_label" binding:"required"` // A, B, C, D, E
	OptionText  string  `json:"option_text"`
	MediaType   string  `json:"media_type"` // NONE, AUDIO, IMAGE
	MediaURL    *string `json:"media_url"`
	SortOrder   int     `json:"sort_order"`
}

// ItemInputDTO represents an input sub-question under a stimulus
type ItemInputDTO struct {
	IDItem            *int64           `json:"id_item"`
	ItemOrder         int              `json:"item_order"`
	QuestionText      string           `json:"question_text" binding:"required"`
	QuestionMediaType string           `json:"question_media_type"` // NONE, AUDIO, IMAGE, VIDEO
	QuestionMediaURL  *string          `json:"question_media_url"`
	ItemType          string           `json:"item_type"` // SINGLE_CHOICE, MULTI_CHOICE
	WeightCorrect     float64          `json:"weight_correct"`
	WeightWrong       float64          `json:"weight_wrong"`
	WeightBlank       float64          `json:"weight_blank"`
	CorrectAnswer     string           `json:"correct_answer" binding:"required"` // "A", "B", etc.
	Explanation       *string          `json:"explanation"`
	Options           []OptionInputDTO `json:"options" binding:"required"`
}

// CreateStimulusDTO represents the payload to create a stimulus with child items
type CreateStimulusDTO struct {
	KodeSoal        string         `json:"kodesoal" binding:"required"`
	StimulusCode    string         `json:"stimulus_code"`
	Title           string         `json:"title" binding:"required"`
	NarrativeText   string         `json:"narrative_text" binding:"required"`
	MediaType       string         `json:"media_type"` // NONE, AUDIO, IMAGE, VIDEO, MULTI
	MediaURL        *string        `json:"media_url"`
	MediaMetadata   map[string]any `json:"media_metadata"`
	CategoryTag     string         `json:"category_tag"`
	DifficultyLevel int            `json:"difficulty_level"` // 1=Easy, 2=Medium, 3=Hard
	SortOrder       int            `json:"sort_order"`
	Items           []ItemInputDTO `json:"items" binding:"required"`
}

// UpdateStimulusDTO represents the payload to update a stimulus
type UpdateStimulusDTO struct {
	Title           string         `json:"title" binding:"required"`
	NarrativeText   string         `json:"narrative_text" binding:"required"`
	MediaType       string         `json:"media_type"`
	MediaURL        *string        `json:"media_url"`
	MediaMetadata   map[string]any `json:"media_metadata"`
	CategoryTag     string         `json:"category_tag"`
	DifficultyLevel int            `json:"difficulty_level"`
	SortOrder       int            `json:"sort_order"`
	Items           []ItemInputDTO `json:"items"`
}

// BlueprintRuleDTO represents a rule inside an exam blueprint
type BlueprintRuleDTO struct {
	CategoryTag      string  `json:"category_tag" binding:"required"`
	DifficultyLevel  int     `json:"difficulty_level"` // 0=Any, 1=Easy, 2=Medium, 3=Hard
	QuotaCount       int     `json:"quota_count" binding:"required"`
	WeightMultiplier float64 `json:"weight_multiplier"`
	SortOrder        int     `json:"sort_order"`
}

// CreateBlueprintDTO represents the payload to create a dynamic blueprint
type CreateBlueprintDTO struct {
	BlueprintCode        string             `json:"blueprint_code" binding:"required"`
	Title                string             `json:"title" binding:"required"`
	Description          *string            `json:"description"`
	TotalTargetQuestions int                `json:"total_target_questions" binding:"required"`
	PassingScore         float64            `json:"passing_score"`
	DurationMinutes      int                `json:"duration_minutes" binding:"required"`
	Rules                []BlueprintRuleDTO `json:"rules" binding:"required"`
}

// SetExamScheduleExtDTO represents the configuration for exam time window and scoring rules
type SetExamScheduleExtDTO struct {
	IDJadwalUjian             int     `json:"idjadwalujian" binding:"required"`
	IDBlueprint               *int64  `json:"id_blueprint"`
	WindowStartTime           string  `json:"window_start_time" binding:"required"` // ISO timestamp or "2026-08-27 08:00:00"
	WindowEndTime             string  `json:"window_end_time" binding:"required"`   // ISO timestamp or "2026-08-27 15:00:00"
	ScoringRule               string  `json:"scoring_rule"`                         // STANDARD, PENALTY_MINUS_ONE, CUSTOM
	DefaultCorrectScore       float64 `json:"default_correct_score"`
	DefaultWrongScore         float64 `json:"default_wrong_score"`
	DefaultBlankScore         float64 `json:"default_blank_score"`
	AutoSubmitOnWindowEnd     bool    `json:"auto_submit_on_window_end"`
	Timezone                  string  `json:"timezone"`
}

// SaveMultiAnswerDTO represents answering a sub-item in a dynamic exam
type SaveMultiAnswerDTO struct {
	ScheduleID      int    `json:"schedule_id" binding:"required"`
	ParticipantCode string `json:"participant_code" binding:"required"`
	IDStimulus      int64  `json:"id_stimulus" binding:"required"`
	IDItem          int64  `json:"id_item" binding:"required"`
	SelectedOption  string `json:"selected_option"` // e.g. "A"
	IsDoubtful      bool   `json:"is_doubtful"`
}
