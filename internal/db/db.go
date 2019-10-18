package db

import (
	"fmt"

	"github.com/antik9/social-net/internal/config"
	"github.com/antik9/social-net/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	Db  *sqlx.DB
	err error
)

func init() {
	connectionParams := fmt.Sprintf(
		"%s:%s@(%s:%s)/%s?%s",
		config.Conf.Database.Username,
		config.Conf.Database.Password,
		config.Conf.Database.Host,
		config.Conf.Database.Port,
		config.Conf.Database.Name,
		config.Conf.Database.Extra,
	)
	Db, err = sqlx.Connect("mysql", connectionParams)
	projecterrors.FailOnErr(err)
}
