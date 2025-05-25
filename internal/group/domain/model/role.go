package model

type Role struct {
	ID   uint   `db:"id"`
	Name string `db:"name"`
}
