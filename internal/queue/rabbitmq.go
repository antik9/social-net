package queue

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"

	"github.com/antik9/social-net/internal/config"
)

type Client struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	messages <-chan amqp.Delivery
	queue    amqp.Queue
}

func NewClient(clientType string) *Client {
	conn, err := amqp.Dial(fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		config.Conf.Rabbit.Username, config.Conf.Rabbit.Password,
		config.Conf.Rabbit.Host, config.Conf.Rabbit.Port,
	))
	if err != nil {
		log.Fatalf("%s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("%s", err)
	}

	q, err := ch.QueueDeclare("feed_message_ids", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("%s", err)
	}
	var messages <-chan amqp.Delivery

	if clientType == "consumer" {
		msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
		if err != nil {
			log.Fatalf("%s", err)
		}
		messages = msgs
	}
	return &Client{
		conn:     conn,
		channel:  ch,
		queue:    q,
		messages: messages,
	}
}

func (c *Client) Close() {
	c.channel.Close()
	c.conn.Close()
}

func (c *Client) SendMessage(message string) {
	err := c.channel.Publish(
		"",           // exchange
		c.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func (c *Client) ReadMessage() string {
	message := <-c.messages
	defer message.Ack(false)
	return string(message.Body)
}
