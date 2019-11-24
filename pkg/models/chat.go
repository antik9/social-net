package models

import (
	"github.com/antik9/social-net/internal/db"
)

type Message struct {
	SenderId  int    `db:"sender_id"`
	Timestamp string `db:"timestamp"`
	Message   string `db:"message"`
}

func orderUsers(user, other *User) (user1Id, user2Id int) {
	user1Id, user2Id = user.Id, other.Id
	if user1Id > user2Id {
		user1Id, user2Id = user2Id, user1Id
	}
	return
}

func GetMessagesForUsers(user, other *User) []Message {
	var messages []Message
	user1Id, user2Id := orderUsers(user, other)
	chatDb := db.GetChatShard(user1Id + user2Id)
	chatDb.Select(
		&messages,
		`SELECT sender_id, timestamp, message FROM message
		WHERE user1_id = ? AND user2_id = ? ORDER BY timestamp`, user1Id, user2Id,
	)
	return messages
}

func SaveMessage(message string, user, other *User) {
	user1Id, user2Id := orderUsers(user, other)
	chatDb := db.GetChatShard(user1Id + user2Id)
	chatDb.MustExec(
		`INSERT INTO message (user1_id, user2_id, sender_id, message)
		VALUES (?, ?, ?, ?)`, user1Id, user2Id, user.Id, message,
	)
}
