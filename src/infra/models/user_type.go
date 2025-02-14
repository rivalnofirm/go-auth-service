package models

type UserType struct {
	Id   int64  `db:"id"`
	Type string `db:"type"`
}
