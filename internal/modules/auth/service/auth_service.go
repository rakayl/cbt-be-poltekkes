package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"poltekkes-cat-backend/internal/modules/auth/dto"
	"poltekkes-cat-backend/internal/modules/auth/repository"
	"poltekkes-cat-backend/internal/shared/jwt"

	"github.com/google/uuid"
)

type AuthService interface {
	LoginAdmin(ctx context.Context, req *dto.AdminLoginRequestDTO) (*dto.AdminLoginResponseDTO, error)
	LoginParticipant(ctx context.Context, req *dto.ParticipantLoginRequestDTO) (*dto.ParticipantLoginResponseDTO, error)
	LoginEPParticipant(ctx context.Context, req *dto.ParticipantLoginRequestDTO) (*dto.ParticipantLoginResponseDTO, error)
	GetModuleMenus(ctx context.Context, moduleID, roleID string) ([]*dto.ModuleMenuResponseDTO, error)
	SwitchModule(ctx context.Context, req *dto.SwitchModuleRequestDTO) (map[string]string, error)
}

type authService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) LoginAdmin(ctx context.Context, req *dto.AdminLoginRequestDTO) (*dto.AdminLoginResponseDTO, error) {
	user, roles, err := s.repo.AuthenticateAdmin(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	duration := 7 * 24 * time.Hour
	if req.RememberMe {
		duration = 365 * 24 * time.Hour
	}

	var roleNames []string
	var roleDTOs []dto.UserRoleDTO
	for _, r := range roles {
		roleNames = append(roleNames, r.RoleID)
		roleDTOs = append(roleDTOs, dto.UserRoleDTO{
			UserID:   r.UserID,
			RoleID:   r.RoleID,
			RoleName: r.RoleName,
			UnitID:   r.UnitID,
			UnitName: r.UnitName,
		})
	}

	token, err := jwt.GenerateAdminToken(user.UserID, user.Username, user.RealName, roleNames, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AdminLoginResponseDTO{
		UserID:      user.UserID,
		Username:    user.Username,
		FullName:    user.RealName,
		Email:       user.Email,
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(duration.Seconds()),
		Roles:       roleDTOs,
	}, nil
}

func (s *authService) LoginParticipant(ctx context.Context, req *dto.ParticipantLoginRequestDTO) (*dto.ParticipantLoginResponseDTO, error) {
	peserta, schedules, err := s.repo.AuthenticateParticipant(ctx, req.ParticipantCode, req.Password)
	if err != nil {
		return nil, err
	}

	newToken := uuid.New().String()
	_ = s.repo.UpdateParticipantToken(ctx, peserta.KodePeserta, newToken)

	duration := 7 * 24 * time.Hour
	jwtToken, err := jwt.GenerateParticipantToken(peserta.KodePeserta, peserta.Nama, newToken, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate participant token: %w", err)
	}

	var schedDTOs []dto.ParticipantSchedDTO
	for _, sc := range schedules {
		var scannedAtStr *string
		if sc.BarcodeScannedAt != nil {
			s := sc.BarcodeScannedAt.Format(time.RFC3339)
			scannedAtStr = &s
		}
		schedDTOs = append(schedDTOs, dto.ParticipantSchedDTO{
			ScheduleID:       sc.IDJadwalUjian,
			ExamID:           sc.IDUjian,
			ExamName:         sc.NamaUjian,
			QuestionBankCode: sc.KodeSoal,
			RoomID:           sc.IDRuang,
			RoomName:         sc.NamaRuang,
			ExamDate:         sc.TglUjian.Format(time.RFC3339),
			EndDate:          sc.TglSelesai.Format(time.RFC3339),
			StartTime:        formatTimeString(sc.JamMulai),
			EndTime:          formatTimeString(sc.JamSelesai),
			DurationMinutes:  sc.WaktuPengerjaan,
			SessionToken:     sc.TokenUjian,
			Capacity:         sc.Kapasitas,
			ScheduleStatus:   sc.ScheduleStatus,
			IsVerified:       sc.IsVerified,
			BarcodeScannedAt: scannedAtStr,
			ExamBarcode:      sc.ExamBarcode,
			HasStarted:       sc.HasStarted,
			IsFinished:       sc.IsFinished,
			RemainingSeconds: sc.RemainingSeconds,
			MaxViolations:    sc.MaxViolations,
			IsEP:             sc.IsEP,
			ExamType:         sc.ExamType,
		})
	}

	return &dto.ParticipantLoginResponseDTO{
		ParticipantCode:  peserta.KodePeserta,
		Name:             peserta.Nama,
		ParticipantToken: jwtToken,
		ActiveSchedules:  schedDTOs,
	}, nil
}

func (s *authService) LoginEPParticipant(ctx context.Context, req *dto.ParticipantLoginRequestDTO) (*dto.ParticipantLoginResponseDTO, error) {
	peserta, schedules, err := s.repo.AuthenticateEPParticipant(ctx, req.ParticipantCode, req.Password)
	if err != nil {
		return nil, err
	}

	newToken := uuid.New().String()
	_ = s.repo.UpdateEPParticipantToken(ctx, peserta.KodePeserta, newToken)

	duration := 7 * 24 * time.Hour
	jwtToken, err := jwt.GenerateParticipantToken(peserta.KodePeserta, peserta.Nama, newToken, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate participant token: %w", err)
	}

	var schedDTOs []dto.ParticipantSchedDTO
	for _, sc := range schedules {
		var scannedAtStr *string
		if sc.BarcodeScannedAt != nil {
			s := sc.BarcodeScannedAt.Format(time.RFC3339)
			scannedAtStr = &s
		}
		schedDTOs = append(schedDTOs, dto.ParticipantSchedDTO{
			ScheduleID:       sc.IDJadwalUjian,
			ExamID:           sc.IDUjian,
			ExamName:         sc.NamaUjian,
			QuestionBankCode: sc.KodeSoal,
			RoomID:           sc.IDRuang,
			RoomName:         sc.NamaRuang,
			ExamDate:         sc.TglUjian.Format(time.RFC3339),
			EndDate:          sc.TglSelesai.Format(time.RFC3339),
			StartTime:        formatTimeString(sc.JamMulai),
			EndTime:          formatTimeString(sc.JamSelesai),
			DurationMinutes:  sc.WaktuPengerjaan,
			SessionToken:     sc.TokenUjian,
			Capacity:         sc.Kapasitas,
			ScheduleStatus:   sc.ScheduleStatus,
			IsVerified:       sc.IsVerified,
			BarcodeScannedAt: scannedAtStr,
			ExamBarcode:      sc.ExamBarcode,
			HasStarted:       sc.HasStarted,
			IsFinished:       sc.IsFinished,
			RemainingSeconds: sc.RemainingSeconds,
			MaxViolations:    sc.MaxViolations,
			IsEP:             sc.IsEP,
			ExamType:         sc.ExamType,
		})
	}

	return &dto.ParticipantLoginResponseDTO{
		ParticipantCode:  peserta.KodePeserta,
		Name:             peserta.Nama,
		ParticipantToken: jwtToken,
		ActiveSchedules:  schedDTOs,
	}, nil
}

func (s *authService) GetModuleMenus(ctx context.Context, moduleID, roleID string) ([]*dto.ModuleMenuResponseDTO, error) {
	flatMenus, err := s.repo.GetModuleMenus(ctx, moduleID, roleID)
	if err != nil {
		return nil, err
	}

	menuMap := make(map[int]*dto.ModuleMenuResponseDTO)
	var rootMenus []*dto.ModuleMenuResponseDTO

	for _, item := range flatMenus {
		dtoItem := &dto.ModuleMenuResponseDTO{
			MenuID:       item.IDMenu,
			ParentMenuID: item.ParentMenu,
			TargetID:     item.IDTarget,
			ModuleID:     item.IDModul,
			Title:        item.NamaMenu,
			Level:        item.LevelMenu,
			Icon:         item.FAIcon,
			TargetFile:   item.NamaFile,
			Submenus:     []*dto.ModuleMenuResponseDTO{},
		}
		menuMap[item.IDMenu] = dtoItem
	}

	for _, item := range flatMenus {
		dtoItem := menuMap[item.IDMenu]
		if item.ParentMenu == nil || *item.ParentMenu == 0 {
			rootMenus = append(rootMenus, dtoItem)
		} else {
			if parent, exists := menuMap[*item.ParentMenu]; exists {
				parent.Submenus = append(parent.Submenus, dtoItem)
			} else {
				rootMenus = append(rootMenus, dtoItem)
			}
		}
	}

	return rootMenus, nil
}

func (s *authService) SwitchModule(ctx context.Context, req *dto.SwitchModuleRequestDTO) (map[string]string, error) {
	return map[string]string{
		"target_module":   req.TargetModule,
		"redirect_url":    "/" + req.TargetModule + "/login",
		"handshake_token": uuid.New().String(),
	}, nil
}

func formatTimeString(t string) string {
	t = strings.TrimSpace(t)
	if len(t) == 4 && !strings.Contains(t, ":") {
		return t[:2] + ":" + t[2:]
	}
	if strings.Contains(t, ":") {
		parts := strings.Split(t, ":")
		if len(parts) >= 2 {
			hh := parts[0]
			if len(hh) == 1 {
				hh = "0" + hh
			}
			mm := parts[1]
			if len(mm) == 1 {
				mm = "0" + mm
			}
			return hh + ":" + mm
		}
	}
	if t == "" {
		return "00:00"
	}
	return t
}
