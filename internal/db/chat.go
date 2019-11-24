package db

var messageSchema = `
CREATE TABLE IF NOT EXISTS message (
	user1_id INT NOT NULL,
	user2_id INT NOT NULL,
	sender_id INT NOT NULL,
	timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
	message VARCHAR(256)
)
`
