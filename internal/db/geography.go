package db

import (
// 	_ "github.com/go-sql-driver/mysql"
// 	"github.com/jmoiron/sqlx"
)

var citySchema = `
CREATE TABLE IF NOT EXISTS city (
	id INT NOT NULL AUTO_INCREMENT,
	name VARCHAR(255) NOT NULL UNIQUE,
	PRIMARY KEY(id)
)
`
