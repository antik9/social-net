package main

import (
	"github.com/antik9/social-net/internal/db"
)

func main() {
	db.InitialMigrate()
}
