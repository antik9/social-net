package models

type City struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}
