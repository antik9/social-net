package models

import (
	"strconv"

	"github.com/tarantool/go-tarantool"

	"github.com/antik9/social-net/internal/db"
	tnt "github.com/antik9/social-net/internal/tarantool"
)

type FeedMessage struct {
	AuthorId   int    `db:"author_id"`
	AuthorName string `db:"author_name"`
	Timestamp  string `db:"timestamp"`
	Message    string `db:"message"`
}

func (u *User) CreateFeedMessage(message string) (int64, error) {
	result := db.Db.MustExec(
		`INSERT INTO feed_message (author_id, message) VALUES (?, ?)`,
		u.Id, message,
	)
	return result.LastInsertId()
}

func GetSubscribersForMessage(messageId int) []int {
	var authorId int

	db.Db.QueryRow(
		`SELECT author_id from feed_message WHERE id = ?`,
		messageId,
	).Scan(&authorId)

	if authorId != 0 {
		var userIds []int
		db.Db.Select(
			&userIds,
			`SELECT subscriber_id FROM subscription WHERE author_id = ?`,
			authorId,
		)
		return userIds
	}
	return make([]int, 0)
}

func SaveUserFeedMessageLink(userId, messageId int) {
	dbShard := db.GetChatShard(userId)
	dbShard.MustExec(
		`INSERT INTO feed_user_link (user_id, feed_message_id) VALUES (?, ?)`,
		userId, messageId,
	)
}

func (u *User) ListOwnFeedLimitBy(limit int) []FeedMessage {
	var messages []FeedMessage

	/*
		replicaDb := db.GetRandomReplica()
		replicaDb.Select(
			&messages,
			`SELECT author_id, ? AS author_name, timestamp, message FROM feed_message
		 	WHERE author_id = ? ORDER BY id DESC LIMIT ?`,
			u.FirstName, u.Id, limit,
		)
	*/
	resp, _ := tnt.Client.Select(
		"feed_message",
		"secondary", 0, uint32(limit), tarantool.IterEq,
		[]interface{}{uint(u.Id)},
	)
	for _, tuple := range resp.Tuples() {
		messages = append(messages, FeedMessage{
			AuthorId:   u.Id,
			AuthorName: u.FirstName,
			Timestamp:  tuple[2].(string),
			Message:    tuple[3].(string),
		})
	}

	return messages
}

func (u *User) ListFeedLimitBy(limit int) []FeedMessage {
	var lastIds []int
	dbShard := db.GetChatShard(u.Id)
	dbShard.Select(
		&lastIds,
		`SELECT feed_message_id FROM feed_user_link WHERE user_id = ? ORDER BY id DESC LIMIT ?`,
		u.Id, limit,
	)

	lastIdsStmt := ""
	for _, id := range lastIds {
		lastIdsStmt += "," + strconv.Itoa(id)
	}
	if lastIdsStmt != "" {
		lastIdsStmt = lastIdsStmt[1:]
		lastIdsStmt = "(" + lastIdsStmt + ")"

		var messages []FeedMessage
		db.Db.Select(
			&messages,
			`SELECT author_id, CONCAT(user.first_name, ' ', user.last_name) author_name,
			timestamp, message FROM feed_message
			INNER JOIN user ON feed_message.author_id = user.id
			WHERE feed_message.id IN `+lastIdsStmt+` ORDER BY feed_message.id DESC LIMIT ?`,
			limit,
		)
		return messages
	}
	return make([]FeedMessage, 0)
}

func (u *User) SubscribeTo(author *User) {
	isSubscribed := 0
	db.Db.QueryRow(
		`SELECT COUNT(1) cnt FROM subscription WHERE author_id = ? AND subscriber_id = ?`,
		author.Id, u.Id,
	).Scan(&isSubscribed)

	if isSubscribed == 0 {
		db.Db.MustExec(
			`INSERT INTO subscription (author_id, subscriber_id)
			VALUES (?, ?)`, author.Id, u.Id,
		)
	}
}
