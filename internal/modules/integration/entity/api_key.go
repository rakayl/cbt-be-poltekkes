package entity

import "time"

type APIKey struct {
	ID          int        `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	APIKey      string     `db:"api_key" json:"api_key"`
	IPWhitelist *string    `db:"ip_whitelist" json:"ip_whitelist"`
	IsActive    bool       `db:"is_active" json:"is_active"`
	Description *string    `db:"description" json:"description"`
	LastUsedAt  *time.Time `db:"last_used_at" json:"last_used_at"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	SoftDelete  string     `db:"softdelete" json:"-"`
}

type APIAccessLog struct {
	ID             int64     `db:"id" json:"id"`
	APIKeyID       *int      `db:"api_key_id" json:"api_key_id"`
	ClientName     string    `db:"client_name" json:"client_name"`
	IPAddress      string    `db:"ip_address" json:"ip_address"`
	Method         string    `db:"method" json:"method"`
	Endpoint       string    `db:"endpoint" json:"endpoint"`
	StatusCode     int       `db:"status_code" json:"status_code"`
	ResponseTimeMS int       `db:"response_time_ms" json:"response_time_ms"`
	UserAgent      *string   `db:"user_agent" json:"user_agent"`
	ErrorMessage   *string   `db:"error_message" json:"error_message"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

