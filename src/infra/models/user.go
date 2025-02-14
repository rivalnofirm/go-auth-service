package models

import (
	"database/sql"
)

type User struct {
	Id        int64        `db:"id"`
	Email     string       `db:"email"`
	Password  string       `db:"password"`
	CreatedAt sql.NullTime `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}
