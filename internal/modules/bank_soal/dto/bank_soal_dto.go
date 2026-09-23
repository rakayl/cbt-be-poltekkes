package dto

import "time"

type QuestionBankResponseDTO struct {
	QuestionBankCode string    `json:"question_bank_code"`
	SubjectName      string    `json:"subject_name"`
	Description      string    `json:"description"`
	TotalQuestions       int       `json:"total_questions"`
	TotalQuestionsStatic int       `json:"total_questions_static"`
	TotalStimuli         int       `json:"total_stimuli"`
	CreatedAt            time.Time `json:"created_at"`
}

type QuestionOptionDTO struct {
	OptionKey  int     `json:"option_key"`
	OptionText string  `json:"option_text"`
	MediaType  string  `json:"media_type,omitempty"` // NONE, AUDIO, IMAGE, VIDEO
	MediaURL   *string `json:"media_url,omitempty"`
}

type QuestionItemResponseDTO struct {
	QuestionBankCode  string              `json:"question_bank_code"`
	QuestionNumber    int                 `json:"question_number"`
	QuestionText      string              `json:"question_text"`
	QuestionMediaType string              `json:"question_media_type,omitempty"` // NONE, AUDIO, IMAGE, VIDEO
	QuestionMediaURL  *string             `json:"question_media_url,omitempty"`
	CorrectOption     int                 `json:"correct_option_key,omitempty"`
	CategoryTag       string              `json:"category_tag,omitempty"`
	Bobot             float64             `json:"bobot"`
	BobotBenar        float64             `json:"bobot_benar"`
	BobotSalah        float64             `json:"bobot_salah"`
	Options           []QuestionOptionDTO `json:"options"`
}

type CreateQuestionBankDTO struct {
	QuestionBankCode string `json:"question_bank_code" binding:"required"`
	QuestionBankName string `json:"question_bank_name" binding:"required"`
	Description      string `json:"description"`
}

type UpdateQuestionBankDTO struct {
	QuestionBankName string `json:"question_bank_name" binding:"required"`
	Description      string `json:"description"`
}

type SaveQuestionItemDTO struct {
	QuestionText      string  `json:"question_text" binding:"required"`
	QuestionMediaType string  `json:"question_media_type"` // NONE, AUDIO, IMAGE, VIDEO
	QuestionMediaURL  *string `json:"question_media_url"`
	OptionA           string  `json:"option_a" binding:"required"`
	OptionAMediaType  string  `json:"option_a_media_type"`
	OptionAMediaURL   *string `json:"option_a_media_url"`
	OptionB           string  `json:"option_b" binding:"required"`
	OptionBMediaType  string  `json:"option_b_media_type"`
	OptionBMediaURL   *string `json:"option_b_media_url"`
	OptionC           string  `json:"option_c"`
	OptionCMediaType  string  `json:"option_c_media_type"`
	OptionCMediaURL   *string `json:"option_c_media_url"`
	OptionD           string  `json:"option_d"`
	OptionDMediaType  string  `json:"option_d_media_type"`
	OptionDMediaURL   *string `json:"option_d_media_url"`
	OptionE           string  `json:"option_e"`
	OptionEMediaType  string  `json:"option_e_media_type"`
	OptionEMediaURL   *string `json:"option_e_media_url"`
	CorrectOptionKey  int     `json:"correct_option_key" binding:"required"`
	CategoryTag       string  `json:"category_tag"`
	Bobot             float64 `json:"bobot"`
	BobotBenar        float64 `json:"bobot_benar"`
	BobotSalah        float64 `json:"bobot_salah"`
}

type CopyQuestionBankDTO struct {
	NewBankCode string `json:"new_bank_code" binding:"required"`
	NewBankName string `json:"new_bank_name" binding:"required"`
}
