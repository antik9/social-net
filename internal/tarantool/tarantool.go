package tarantool

import (
	"log"
	"time"

	"github.com/tarantool/go-tarantool"

	"github.com/antik9/social-net/internal/config"
)

var (
	Client *tarantool.Connection
)

func init() {
	var err error

	server := config.Conf.Tarantool.Host
	opts := tarantool.Opts{
		Timeout:       500 * time.Millisecond,
		Reconnect:     1 * time.Second,
		MaxReconnects: 3,
		User:          config.Conf.Tarantool.Username,
		Pass:          config.Conf.Tarantool.Password,
	}

	Client, err = tarantool.Connect(server, opts)
	if err != nil {
		log.Fatal(err)
	}
}
