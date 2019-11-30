package main

import (
	"strconv"

	"github.com/antik9/social-net/internal/queue"
	"github.com/antik9/social-net/pkg/models"
)

func main() {
	consumer := queue.NewClient("consumer")

	for {
		message := consumer.ReadMessage()
		messageId, err := strconv.Atoi(message)
		if err == nil {
			for _, subscriberId := range models.GetSubscribersForMessage(messageId) {
				models.SaveUserFeedMessageLink(subscriberId, messageId)
			}
		}
	}
}
