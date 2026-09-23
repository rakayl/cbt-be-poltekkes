package dto

type AdminLoginRequestDTO struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

type AdminLoginResponseDTO struct {
	UserID      int           `json:"user_id"`
	Username    string        `json:"username"`
	FullName    string        `json:"full_name"`
	Email       string        `json:"email"`
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
	ExpiresIn   int           `json:"expires_in"`
	Roles       []UserRoleDTO `json:"roles"`
}

type UserRoleDTO struct {
	UserID   int    `json:"user_id"`
	RoleID   string `json:"role_id"`
	RoleName string `json:"role_name"`
	UnitID   string `json:"unit_id"`
	UnitName string `json:"unit_name"`
}

type ParticipantLoginRequestDTO struct {
	ParticipantCode string `json:"participant_code" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

type ParticipantLoginResponseDTO struct {
	ParticipantCode  string                `json:"participant_code"`
	Name             string                `json:"name"`
	ParticipantToken string                `json:"participant_token"`
	ActiveSchedules  []ParticipantSchedDTO `json:"active_schedules"`
}

type ParticipantSchedDTO struct {
	ScheduleID       int    `json:"schedule_id"`
	ExamID           int    `json:"exam_id"`
	ExamName         string `json:"exam_name"`
	QuestionBankCode string `json:"question_bank_code"`
	RoomID           int    `json:"room_id"`
	RoomName         string `json:"room_name"`
	ExamDate         string `json:"exam_date"`
	EndDate          string `json:"end_date"`
	StartTime        string `json:"start_time"`
	EndTime          string `json:"end_time"`
	DurationMinutes  int    `json:"duration_minutes"`
	SessionToken     string `json:"session_token"`
	Capacity         int     `json:"capacity"`
	ScheduleStatus   string  `json:"schedule_status"`
	IsVerified       int     `json:"is_verified"`
	BarcodeScannedAt *string `json:"barcode_scanned_at,omitempty"`
	ExamBarcode      string  `json:"exam_barcode"`
	HasStarted       bool    `json:"has_started"`
	IsFinished       bool    `json:"is_finished"`
	RemainingSeconds int     `json:"remaining_seconds"`
	MaxViolations    int     `json:"max_violations"`
	IsEP             bool    `json:"is_ep"`
	ExamType         string  `json:"exam_type"`
}

type SwitchModuleRequestDTO struct {
	TargetModule string `json:"target_module" binding:"required"`
	RoleID       string `json:"role_id" binding:"required"`
	UnitID       string `json:"unit_id" binding:"required"`
}

type ModuleMenuResponseDTO struct {
	MenuID       int                      `json:"menu_id"`
	ParentMenuID *int                     `json:"parent_menu_id,omitempty"`
	TargetID     *string                  `json:"target_id,omitempty"`
	ModuleID     string                   `json:"module_id"`
	Title        string                   `json:"title"`
	Level        int                      `json:"level"`
	Icon         *string                  `json:"icon,omitempty"`
	TargetFile   *string                  `json:"target_file,omitempty"`
	Submenus     []*ModuleMenuResponseDTO `json:"submenus,omitempty"`
}
