package db

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/antik9/social-net/internal/config"
	"github.com/antik9/social-net/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	Db       *sqlx.DB
	chatDbs  []*sqlx.DB
	replicas []*sqlx.DB
	err      error
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

	for _, replica := range config.Conf.Database.Replicas {
		connectionParams = fmt.Sprintf(
			"%s:%s@(%s)/%s?%s",
			config.Conf.Database.Username,
			config.Conf.Database.Password,
			replica,
			config.Conf.Database.Name,
			config.Conf.Database.Extra,
		)
		db, err := sqlx.Connect("mysql", connectionParams)
		projecterrors.FailOnErr(err)
		replicas = append(replicas, db)
	}

	for _, shard := range config.Conf.Database.ChatShards {
		connectionParams = fmt.Sprintf(
			"%s:%s@(%s)/%s?%s",
			config.Conf.Database.Username,
			config.Conf.Database.Password,
			shard,
			config.Conf.Database.Name,
			config.Conf.Database.Extra,
		)
		db, err := sqlx.Connect("mysql", connectionParams)
		projecterrors.FailOnErr(err)
		chatDbs = append(chatDbs, db)
	}
}

func GetRandomReplica() *sqlx.DB {
	if len(replicas) > 0 {
		rand.Seed(time.Now().Unix())
		return replicas[rand.Intn(len(replicas))]
	}
	return Db
}

func GetChatShard(hash int) *sqlx.DB {
	if len(chatDbs) == 0 {
		log.Fatal("no chat shards available")
	}
	return chatDbs[hash%len(chatDbs)]
}
