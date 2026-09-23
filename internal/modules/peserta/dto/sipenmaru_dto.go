package dto

import (
	"fmt"
	"strings"
)

// SipenmaruFilterOptionsDTO memuat seluruh opsi filter untuk modal impor Sipenmaru
type SipenmaruFilterOptionsDTO struct {
	CBTPeriods    []CBTPeriodOptionDTO    `json:"cbt_periods"`
	PMBPeriods    []string                `json:"pmb_periods"`
	SistemKuliah  []SistemKuliahOptionDTO `json:"sistem_kuliah"`
	JalurDaftar   []JalurOptionDTO        `json:"jalur_daftar"`
	GelombangList []GelombangOptionDTO    `json:"gelombang_list"`
}

type CBTPeriodOptionDTO struct {
	IDPeriode   int    `json:"id_periode"`
	NamaPeriode string `json:"nama_periode"`
}

type SistemKuliahOptionDTO struct {
	IDSistemKuliah   int    `json:"id_sistem_kuliah"`
	NamaSistemKuliah string `json:"nama_sistem_kuliah"`
}

type JalurOptionDTO struct {
	IDJalurPendaftaran   int    `json:"id_jalur_pendaftaran"`
	NamaJalurPendaftaran string `json:"nama_jalur_pendaftaran"`
}

type GelombangOptionDTO struct {
	IDGelombang   int    `json:"id_gelombang"`
	NamaGelombang string `json:"nama_gelombang"`
}

// SipenmaruCandidateDTO adalah representasi pendaftar yang siap di-preview
type SipenmaruCandidateDTO struct {
	IDPendaftar          string  `json:"id_pendaftar"`
	Nama                 string  `json:"nama"`
	JK                   string  `json:"jk"`
	HP                   string  `json:"hp"`
	Email                string  `json:"email"`
	Alamat               string  `json:"alamat"`
	IDKota               *int    `json:"id_kota"`
	IDPeriode            string  `json:"id_periode"`
	IDSistemKuliah       int     `json:"id_sistem_kuliah"`
	NamaSistemKuliah     string  `json:"nama_sistem_kuliah"`
	IDJalurPendaftaran   int     `json:"id_jalur_pendaftaran"`
	NamaJalurPendaftaran string  `json:"nama_jalur_pendaftaran"`
	IDGelombang          int     `json:"id_gelombang"`
	NamaGelombang        string  `json:"nama_gelombang"`
	IDUnit1              *string `json:"id_unit1"`
	NamaUnit1            string  `json:"nama_unit1"`
	IDUnit2              *string `json:"id_unit2"`
	NamaUnit2            string  `json:"nama_unit2"`
	IsAdministrasi       int     `json:"is_administrasi"`
	NoUjianSipenmaru     string  `json:"no_ujian_sipenmaru"`
	IsImported           bool    `json:"is_imported"`
	ExistingKodePeserta  string  `json:"existing_kode_peserta,omitempty"`
}

// SipenmaruPreviewRequestDTO adalah kriteria filter pendaftar untuk preview
type SipenmaruPreviewRequestDTO struct {
	CBTPeriodID        int         `json:"cbt_period_id"`
	PMBPeriodID        interface{} `json:"pmb_period_id"`
	IDSistemKuliah     int         `json:"id_sistem_kuliah"`
	IDJalurPendaftaran int         `json:"id_jalur_pendaftaran"`
	IDGelombang        int         `json:"id_gelombang"`
	OnlyAdministrasi   bool        `json:"only_administrasi"`
}

func (r *SipenmaruPreviewRequestDTO) GetPMBPeriodString() string {
	return parsePeriodString(r.PMBPeriodID)
}

// SipenmaruPreviewResponseDTO hasil ringkasan dan daftar kandidat
type SipenmaruPreviewResponseDTO struct {
	TotalCount     int                     `json:"total_count"`
	EligibleCount  int                     `json:"eligible_count"`
	ImportedCount  int                     `json:"imported_count"`
	ReadyCount     int                     `json:"ready_count"`
	Candidates     []SipenmaruCandidateDTO `json:"candidates"`
}

// ImportSipenmaruRequestDTO instruksi eksekusi impor
type ImportSipenmaruRequestDTO struct {
	CBTPeriodID          int         `json:"cbt_period_id" binding:"required"`
	PMBPeriodID          interface{} `json:"pmb_period_id"`
	IDSistemKuliah       int         `json:"id_sistem_kuliah"`
	IDJalurPendaftaran   int         `json:"id_jalur_pendaftaran"`
	IDGelombang          int         `json:"id_gelombang"`
	OnlyAdministrasi     bool        `json:"only_administrasi"`
	SelectedPendaftarIDs []string    `json:"selected_pendaftar_ids"` // opsional: jika spesifik memilih
}

func (r *ImportSipenmaruRequestDTO) GetPMBPeriodString() string {
	return parsePeriodString(r.PMBPeriodID)
}

func parsePeriodString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case float64:
		return fmt.Sprintf("%.0f", val)
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// ImportedParticipantItem detail peserta yang sukses diimpor
type ImportedParticipantItem struct {
	KodePeserta string `json:"kode_peserta"`
	IDPendaftar string `json:"id_pendaftar"`
	Nama        string `json:"nama"`
}

// ImportSipenmaruResponseDTO laporan hasil eksekusi impor
type ImportSipenmaruResponseDTO struct {
	TotalProcessed       int                       `json:"total_processed"`
	ImportedCount        int                       `json:"imported_count"`
	SkippedCount         int                       `json:"skipped_count"`
	ErrorCount           int                       `json:"error_count"`
	Errors               []string                  `json:"errors"`
	ImportedParticipants []ImportedParticipantItem `json:"imported_participants"`
}

// ExamDirectImportSipenmaruRequestDTO permintaan impor langsung pada Detail Ujian
type ExamDirectImportSipenmaruRequestDTO struct {
	ImportSipenmaruRequestDTO
	AutoPlotting bool `json:"auto_plotting"` // langsung plotting ke sesi & ruang yang kosong
}

// ExamDirectImportSipenmaruResponseDTO respon impor langsung pada ujian
type ExamDirectImportSipenmaruResponseDTO struct {
	ImportSipenmaruResponseDTO
	ExamRegisteredCount int `json:"exam_registered_count"`
	SessionPlottedCount int `json:"session_plotted_count"`
}
