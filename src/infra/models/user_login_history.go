package models

import (
	"database/sql"
)

type UserLoginHistory struct {
	Id           int64          `db:"id"`
	UserId       int64          `db:"user_id"`
	LoginTime    sql.NullTime   `db:"login_time"`
	IpAddress    sql.NullString `db:"ip_address"`
	UserAgent    sql.NullString `db:"user_agent"`
	LogoutTime   sql.NullTime   `db:"logout_time"`
	LogoutReason sql.NullString `db:"logout_reason"`
}
