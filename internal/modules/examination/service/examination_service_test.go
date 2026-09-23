package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"poltekkes-cat-backend/internal/modules/examination/dto"
	"poltekkes-cat-backend/internal/modules/examination/entity"
	"poltekkes-cat-backend/internal/modules/examination/repository"
	"poltekkes-cat-backend/internal/modules/examination/service"
)

// mockExaminationRepository embeds repository.ExaminationRepository
type mockExaminationRepository struct {
	repository.ExaminationRepository
	lastRecordedEvent *entity.SecurityEvent
	mockResponse      *dto.SecurityEventResponseDTO
	mockError         error
}

func (m *mockExaminationRepository) RecordSecurityEvent(ctx context.Context, event *entity.SecurityEvent) (*dto.SecurityEventResponseDTO, error) {
	m.lastRecordedEvent = event
	if m.mockResponse != nil {
		return m.mockResponse, m.mockError
	}
	return &dto.SecurityEventResponseDTO{
		CurrentRiskScore:    event.RiskScore,
		RiskLevel:           "LOW",
		TabSwitchCount:      1,
		FullscreenExitCount: 0,
		IsLocked:            0,
		AutoLocked:          false,
		MaxViolations:       5,
	}, m.mockError
}

func TestRecordSecurityEvent_RiskScoring(t *testing.T) {
	testCases := []struct {
		eventType        string
		expectedScore    int
		expectedSeverity string
	}{
		{"TAB_SWITCH", 10, "MEDIUM"},
		{"FULLSCREEN_EXIT", 10, "MEDIUM"},
		{"FULLSCREEN_REQUIRED", 10, "MEDIUM"},
		{"WINDOW_BLUR", 5, "LOW"},
		{"COPY_ATTEMPT", 5, "MEDIUM"},
		{"PASTE_ATTEMPT", 5, "MEDIUM"},
		{"CUT_ATTEMPT", 5, "MEDIUM"},
		{"PRINT_ATTEMPT", 15, "HIGH"},
		{"DEVTOOLS_ATTEMPT", 15, "HIGH"},
		{"FORBIDDEN_SHORTCUT", 15, "HIGH"},
		{"CAMERA_DISCONNECTED", 15, "HIGH"},
		{"NETWORK_DISCONNECTED", 10, "MEDIUM"},
		{"LONG_IDLE", 10, "MEDIUM"},
		{"DEVICE_CHANGE", 25, "HIGH"},
		{"CONCURRENT_LOGIN", 30, "CRITICAL"},
		{"UNKNOWN_EVENT", 5, "LOW"},
	}

	for _, tc := range testCases {
		t.Run(tc.eventType, func(t *testing.T) {
			mockRepo := &mockExaminationRepository{}
			svc := service.NewExaminationService(mockRepo)

			req := &dto.SecurityEventRequestDTO{
				ScheduleID: 1,
				EventType:  tc.eventType,
				DeviceID:   "device-123",
				Metadata: map[string]interface{}{
					"test": "val",
				},
			}

			resp, err := svc.RecordSecurityEvent(context.Background(), "PESERTA001", "127.0.0.1", "Mozilla/5.0", req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("expected non-nil response")
			}

			if mockRepo.lastRecordedEvent == nil {
				t.Fatal("expected lastRecordedEvent to be populated")
			}

			ev := mockRepo.lastRecordedEvent
			if ev.RiskScore != tc.expectedScore {
				t.Errorf("[%s] expected risk score %d, got %d", tc.eventType, tc.expectedScore, ev.RiskScore)
			}
			if ev.Severity != tc.expectedSeverity {
				t.Errorf("[%s] expected severity %s, got %s", tc.eventType, tc.expectedSeverity, ev.Severity)
			}
			if ev.KodePeserta != "PESERTA001" {
				t.Errorf("expected KodePeserta 'PESERTA001', got '%s'", ev.KodePeserta)
			}
			if ev.IDJadwal != 1 {
				t.Errorf("expected IDJadwal 1, got %d", ev.IDJadwal)
			}
			if ev.DeviceID == nil || *ev.DeviceID != "device-123" {
				t.Errorf("expected DeviceID 'device-123', got '%v'", ev.DeviceID)
			}
			if ev.Metadata == nil {
				t.Error("expected metadata to be serialized")
			} else {
				var meta map[string]interface{}
				if err := json.Unmarshal([]byte(*ev.Metadata), &meta); err != nil {
					t.Errorf("failed to unmarshal metadata: %v", err)
				}
				if meta["test"] != "val" {
					t.Errorf("expected metadata test=val, got %v", meta["test"])
				}
			}
		})
	}
}

func TestRecordSecurityEvent_AutoLockThresholdResponse(t *testing.T) {
	mockRepo := &mockExaminationRepository{
		mockResponse: &dto.SecurityEventResponseDTO{
			CurrentRiskScore:    85,
			RiskLevel:           "HIGH",
			TabSwitchCount:      5,
			FullscreenExitCount: 2,
			IsLocked:            1,
			AutoLocked:          true,
			MaxViolations:       5,
		},
	}
	svc := service.NewExaminationService(mockRepo)

	req := &dto.SecurityEventRequestDTO{
		ScheduleID: 10,
		EventType:  "TAB_SWITCH",
		DeviceID:   "dev-abc",
	}

	resp, err := svc.RecordSecurityEvent(context.Background(), "PESERTA002", "192.168.1.50", "Chrome", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.AutoLocked {
		t.Error("expected AutoLocked to be true")
	}
	if resp.IsLocked != 1 {
		t.Errorf("expected IsLocked 1, got %d", resp.IsLocked)
	}
	if resp.TabSwitchCount != 5 {
		t.Errorf("expected TabSwitchCount 5, got %d", resp.TabSwitchCount)
	}
	if resp.MaxViolations != 5 {
		t.Errorf("expected MaxViolations 5, got %d", resp.MaxViolations)
	}
	if resp.CurrentRiskScore != 85 {
		t.Errorf("expected CurrentRiskScore 85, got %d", resp.CurrentRiskScore)
	}
}
