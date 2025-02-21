package models

import "database/sql"

type UserRefreshToken struct {
	Id               int64        `db:"id"`
	UserId           int64        `db:"user_id"`
	RefreshTokenHash string       `db:"refresh_token_hash"`
	ExpiresAt        sql.NullTime `db:"expires_at"`
	UserAgent        string       `db:"user_agent"`
	IsActive         bool         `db:"is_active"`
}
