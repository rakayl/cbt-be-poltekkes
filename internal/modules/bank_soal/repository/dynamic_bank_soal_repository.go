package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"poltekkes-cat-backend/internal/modules/bank_soal/dto"
	"poltekkes-cat-backend/internal/modules/bank_soal/entity"

	"github.com/jmoiron/sqlx"
)

type DynamicBankSoalRepository interface {
	GetStimuliByBankCode(ctx context.Context, bankCode string) ([]*entity.BankStimulus, error)
	GetStimulusByID(ctx context.Context, id int64) (*entity.BankStimulus, error)
	CreateStimulus(ctx context.Context, req *dto.CreateStimulusDTO) (*entity.BankStimulus, error)
	UpdateStimulus(ctx context.Context, id int64, req *dto.UpdateStimulusDTO) error
	DeleteStimulus(ctx context.Context, id int64) error

	// Blueprint Methods
	GetBlueprints(ctx context.Context) ([]*entity.ExamBlueprint, error)
	GetBlueprintByID(ctx context.Context, id int64) (*entity.ExamBlueprint, error)
	CreateBlueprint(ctx context.Context, req *dto.CreateBlueprintDTO) (*entity.ExamBlueprint, error)

	// Exam Schedule Extension & Time Window Methods
	GetExamScheduleExt(ctx context.Context, scheduleID int) (*entity.ExamScheduleExt, error)
	SetExamScheduleExt(ctx context.Context, req *dto.SetExamScheduleExtDTO) error
}

type dynamicBankSoalRepository struct {
	db *sqlx.DB
}

func NewDynamicBankSoalRepository(db *sqlx.DB) DynamicBankSoalRepository {
	return &dynamicBankSoalRepository{db: db}
}

func (r *dynamicBankSoalRepository) GetStimuliByBankCode(ctx context.Context, bankCode string) ([]*entity.BankStimulus, error) {
	query := `
		SELECT id_stimulus, kodesoal, stimulus_code, title, narrative_text, media_type, media_url,
		       media_metadata, category_tag, difficulty_level, sort_order, softdelete, created_at, updated_at
		FROM cat.cat_bank_stimulus
		WHERE kodesoal = $1 AND (softdelete = '0' OR softdelete IS NULL)
		ORDER BY sort_order ASC, id_stimulus ASC
	`
	var stimuli []*entity.BankStimulus
	err := r.db.SelectContext(ctx, &stimuli, query, bankCode)
	if err != nil {
		return nil, err
	}

	for _, s := range stimuli {
		items, err := r.getItemsByStimulusID(ctx, s.IDStimulus)
		if err == nil {
			s.Items = items
		}
	}

	return stimuli, nil
}

func (r *dynamicBankSoalRepository) getItemsByStimulusID(ctx context.Context, stimulusID int64) ([]*entity.StimulusItem, error) {
	query := `
		SELECT id_item, id_stimulus, item_order, question_text, question_media_type, question_media_url,
		       item_type, weight_correct, weight_wrong, weight_blank, correct_answer, explanation,
		       softdelete, created_at, updated_at
		FROM cat.cat_stimulus_items
		WHERE id_stimulus = $1 AND (softdelete = '0' OR softdelete IS NULL)
		ORDER BY item_order ASC, id_item ASC
	`
	var items []*entity.StimulusItem
	err := r.db.SelectContext(ctx, &items, query, stimulusID)
	if err != nil {
		return nil, err
	}

	for _, it := range items {
		opts, err := r.getOptionsByItemID(ctx, it.IDItem)
		if err == nil {
			it.Options = opts
		}
	}

	return items, nil
}

func (r *dynamicBankSoalRepository) getOptionsByItemID(ctx context.Context, itemID int64) ([]*entity.ItemOption, error) {
	query := `
		SELECT id_option, id_item, option_label, option_text, media_type, media_url, sort_order, created_at, updated_at
		FROM cat.cat_item_options
		WHERE id_item = $1
		ORDER BY sort_order ASC, option_label ASC
	`
	var opts []*entity.ItemOption
	err := r.db.SelectContext(ctx, &opts, query, itemID)
	return opts, err
}

func (r *dynamicBankSoalRepository) GetStimulusByID(ctx context.Context, id int64) (*entity.BankStimulus, error) {
	query := `
		SELECT id_stimulus, kodesoal, stimulus_code, title, narrative_text, media_type, media_url,
		       media_metadata, category_tag, difficulty_level, sort_order, softdelete, created_at, updated_at
		FROM cat.cat_bank_stimulus
		WHERE id_stimulus = $1 AND (softdelete = '0' OR softdelete IS NULL)
	`
	var s entity.BankStimulus
	err := r.db.GetContext(ctx, &s, query, id)
	if err != nil {
		return nil, err
	}

	items, err := r.getItemsByStimulusID(ctx, s.IDStimulus)
	if err == nil {
		s.Items = items
	}
	return &s, nil
}

func (r *dynamicBankSoalRepository) CreateStimulus(ctx context.Context, req *dto.CreateStimulusDTO) (*entity.BankStimulus, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	code := req.StimulusCode
	if code == "" {
		code = fmt.Sprintf("STIM-%d", time.Now().UnixNano()/1e6)
	}
	metaJSON, _ := json.Marshal(req.MediaMetadata)

	var idStimulus int64
	insertStimulusQ := `
		INSERT INTO cat.cat_bank_stimulus (
			kodesoal, stimulus_code, title, narrative_text, media_type, media_url,
			media_metadata, category_tag, difficulty_level, sort_order, softdelete
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, '0')
		RETURNING id_stimulus
	`
	err = tx.QueryRowContext(ctx, insertStimulusQ,
		req.KodeSoal, code, req.Title, req.NarrativeText, req.MediaType, req.MediaURL,
		metaJSON, req.CategoryTag, req.DifficultyLevel, req.SortOrder,
	).Scan(&idStimulus)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat stimulus: %w", err)
	}

	// Insert child items & options
	for _, item := range req.Items {
		var idItem int64
		insertItemQ := `
			INSERT INTO cat.cat_stimulus_items (
				id_stimulus, item_order, question_text, question_media_type, question_media_url,
				item_type, weight_correct, weight_wrong, weight_blank, correct_answer, explanation, softdelete
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, '0')
			RETURNING id_item
		`
		err = tx.QueryRowContext(ctx, insertItemQ,
			idStimulus, item.ItemOrder, item.QuestionText, item.QuestionMediaType, item.QuestionMediaURL,
			item.ItemType, item.WeightCorrect, item.WeightWrong, item.WeightBlank, item.CorrectAnswer, item.Explanation,
		).Scan(&idItem)
		if err != nil {
			return nil, fmt.Errorf("gagal membuat sub-pertanyaan: %w", err)
		}

		for _, opt := range item.Options {
			insertOptQ := `
				INSERT INTO cat.cat_item_options (id_item, option_label, option_text, media_type, media_url, sort_order)
				VALUES ($1, $2, $3, $4, $5, $6)
			`
			_, err = tx.ExecContext(ctx, insertOptQ, idItem, opt.OptionLabel, opt.OptionText, opt.MediaType, opt.MediaURL, opt.SortOrder)
			if err != nil {
				return nil, fmt.Errorf("gagal membuat opsi pilihan: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetStimulusByID(ctx, idStimulus)
}

func (r *dynamicBankSoalRepository) UpdateStimulus(ctx context.Context, id int64, req *dto.UpdateStimulusDTO) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	metaJSON, _ := json.Marshal(req.MediaMetadata)
	updateStimulusQ := `
		UPDATE cat.cat_bank_stimulus
		SET title = $1, narrative_text = $2, media_type = $3, media_url = $4,
		    media_metadata = $5, category_tag = $6, difficulty_level = $7, sort_order = $8, updated_at = NOW()
		WHERE id_stimulus = $9
	`
	_, err = tx.ExecContext(ctx, updateStimulusQ,
		req.Title, req.NarrativeText, req.MediaType, req.MediaURL,
		metaJSON, req.CategoryTag, req.DifficultyLevel, req.SortOrder, id,
	)
	if err != nil {
		return err
	}

	if len(req.Items) > 0 {
		// Re-sync items
		_, _ = tx.ExecContext(ctx, `DELETE FROM cat.cat_stimulus_items WHERE id_stimulus = $1`, id)
		for _, item := range req.Items {
			var idItem int64
			insertItemQ := `
				INSERT INTO cat.cat_stimulus_items (
					id_stimulus, item_order, question_text, question_media_type, question_media_url,
					item_type, weight_correct, weight_wrong, weight_blank, correct_answer, explanation, softdelete
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, '0')
				RETURNING id_item
			`
			err = tx.QueryRowContext(ctx, insertItemQ,
				id, item.ItemOrder, item.QuestionText, item.QuestionMediaType, item.QuestionMediaURL,
				item.ItemType, item.WeightCorrect, item.WeightWrong, item.WeightBlank, item.CorrectAnswer, item.Explanation,
			).Scan(&idItem)
			if err != nil {
				return err
			}

			for _, opt := range item.Options {
				insertOptQ := `
					INSERT INTO cat.cat_item_options (id_item, option_label, option_text, media_type, media_url, sort_order)
					VALUES ($1, $2, $3, $4, $5, $6)
				`
				_, err = tx.ExecContext(ctx, insertOptQ, idItem, opt.OptionLabel, opt.OptionText, opt.MediaType, opt.MediaURL, opt.SortOrder)
				if err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}

func (r *dynamicBankSoalRepository) DeleteStimulus(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE cat.cat_bank_stimulus SET softdelete = '1', updated_at = NOW() WHERE id_stimulus = $1`, id)
	return err
}

func (r *dynamicBankSoalRepository) GetBlueprints(ctx context.Context) ([]*entity.ExamBlueprint, error) {
	query := `
		SELECT id_blueprint, blueprint_code, title, description, total_target_questions, passing_score,
		       duration_minutes, softdelete, created_at, updated_at
		FROM cat.cat_exam_blueprints
		WHERE softdelete = '0' OR softdelete IS NULL
		ORDER BY id_blueprint DESC
	`
	var blueprints []*entity.ExamBlueprint
	err := r.db.SelectContext(ctx, &blueprints, query)
	if err != nil {
		return nil, err
	}

	for _, bp := range blueprints {
		var rules []*entity.BlueprintRule
		_ = r.db.SelectContext(ctx, &rules, `SELECT * FROM cat.cat_blueprint_rules WHERE id_blueprint = $1 ORDER BY sort_order ASC`, bp.IDBlueprint)
		bp.Rules = rules
	}

	return blueprints, nil
}

func (r *dynamicBankSoalRepository) GetBlueprintByID(ctx context.Context, id int64) (*entity.ExamBlueprint, error) {
	query := `
		SELECT id_blueprint, blueprint_code, title, description, total_target_questions, passing_score,
		       duration_minutes, softdelete, created_at, updated_at
		FROM cat.cat_exam_blueprints
		WHERE id_blueprint = $1 AND (softdelete = '0' OR softdelete IS NULL)
	`
	var bp entity.ExamBlueprint
	err := r.db.GetContext(ctx, &bp, query, id)
	if err != nil {
		return nil, err
	}

	var rules []*entity.BlueprintRule
	_ = r.db.SelectContext(ctx, &rules, `SELECT * FROM cat.cat_blueprint_rules WHERE id_blueprint = $1 ORDER BY sort_order ASC`, bp.IDBlueprint)
	bp.Rules = rules

	return &bp, nil
}

func (r *dynamicBankSoalRepository) CreateBlueprint(ctx context.Context, req *dto.CreateBlueprintDTO) (*entity.ExamBlueprint, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var idBP int64
	insertBP := `
		INSERT INTO cat.cat_exam_blueprints (blueprint_code, title, description, total_target_questions, passing_score, duration_minutes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id_blueprint
	`
	err = tx.QueryRowContext(ctx, insertBP, req.BlueprintCode, req.Title, req.Description, req.TotalTargetQuestions, req.PassingScore, req.DurationMinutes).Scan(&idBP)
	if err != nil {
		return nil, err
	}

	for _, rule := range req.Rules {
		insertRule := `
			INSERT INTO cat.cat_blueprint_rules (id_blueprint, category_tag, difficulty_level, quota_count, weight_multiplier, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.ExecContext(ctx, insertRule, idBP, rule.CategoryTag, rule.DifficultyLevel, rule.QuotaCount, rule.WeightMultiplier, rule.SortOrder)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetBlueprintByID(ctx, idBP)
}

func (r *dynamicBankSoalRepository) GetExamScheduleExt(ctx context.Context, scheduleID int) (*entity.ExamScheduleExt, error) {
	query := `
		SELECT idjadwalujian, id_blueprint, window_start_time, window_end_time, scoring_rule,
		       default_correct_score, default_wrong_score, default_blank_score,
		       auto_submit_on_window_end, timezone, created_at, updated_at
		FROM cat.cat_exam_schedules_ext
		WHERE idjadwalujian = $1
	`
	var ext entity.ExamScheduleExt
	err := r.db.GetContext(ctx, &ext, query, scheduleID)
	if err != nil {
		return nil, err
	}
	return &ext, nil
}

func (r *dynamicBankSoalRepository) SetExamScheduleExt(ctx context.Context, req *dto.SetExamScheduleExtDTO) error {
	tz := req.Timezone
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	rule := req.ScoringRule
	if rule == "" {
		rule = "STANDARD"
	}

	query := `
		INSERT INTO cat.cat_exam_schedules_ext (
			idjadwalujian, id_blueprint, window_start_time, window_end_time, scoring_rule,
			default_correct_score, default_wrong_score, default_blank_score,
			auto_submit_on_window_end, timezone, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (idjadwalujian)
		DO UPDATE SET
			id_blueprint = EXCLUDED.id_blueprint,
			window_start_time = EXCLUDED.window_start_time,
			window_end_time = EXCLUDED.window_end_time,
			scoring_rule = EXCLUDED.scoring_rule,
			default_correct_score = EXCLUDED.default_correct_score,
			default_wrong_score = EXCLUDED.default_wrong_score,
			default_blank_score = EXCLUDED.default_blank_score,
			auto_submit_on_window_end = EXCLUDED.auto_submit_on_window_end,
			timezone = EXCLUDED.timezone,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		req.IDJadwalUjian, req.IDBlueprint, req.WindowStartTime, req.WindowEndTime, rule,
		req.DefaultCorrectScore, req.DefaultWrongScore, req.DefaultBlankScore,
		req.AutoSubmitOnWindowEnd, tz,
	)
	return err
}
