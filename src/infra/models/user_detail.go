package models

import (
	"database/sql"
)

type UserDetail struct {
	Id         int64          `db:"id"`
	UserId     int64          `db:"user_id"`
	UserTypeId int64          `db:"user_type_id"`
	Phone      sql.NullString `db:"phone"`
	FirstName  sql.NullString `db:"first_name"`
	LastName   sql.NullString `db:"last_name"`
	Picture    sql.NullString `db:"picture"`
	BirthDate  sql.NullTime   `db:"birth_date"`
	Gender     sql.NullString `db:"gender"`
	Verified   string         `db:"verified"`
	CreatedAt  sql.NullTime   `db:"created_at"`
	UpdatedAt  sql.NullTime   `db:"updated_at"`
	DeletedAt  sql.NullTime   `db:"deleted_at"`
}
