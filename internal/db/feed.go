package db

var feedMessageSchema = `
CREATE TABLE IF NOT EXISTS feed_message (
	id INT NOT NULL AUTO_INCREMENT,
	author_id INT NOT NULL,
	timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
	message VARCHAR(256),
	PRIMARY KEY(id),
	FOREIGN KEY (author_id)
		REFERENCES user(id)
		ON DELETE CASCADE
)
`

var feedUserLinkSchema = `
CREATE TABLE IF NOT EXISTS feed_user_link (
	id INT NOT NULL AUTO_INCREMENT,
	user_id INT NOT NULL,
	feed_message_id INT NOT NULL,
	PRIMARY KEY(id)
)
`

var subscriptionSchema = `
CREATE TABLE IF NOT EXISTS subscription (
	subscriber_id INT NOT NULL,
	author_id INT NOT NULL,
	FOREIGN KEY (author_id)
		REFERENCES user(id)
		ON DELETE CASCADE,
	FOREIGN KEY (subscriber_id)
		REFERENCES user(id)
		ON DELETE CASCADE,
	UNIQUE(subscriber_id, author_id)
)
`
