package db

var userSchema = `
CREATE TABLE IF NOT EXISTS user (
	id INT NOT NULL AUTO_INCREMENT,
	first_name VARCHAR(255) NOT NULL,
	last_name VARCHAR(255) NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	password VARCHAR(128) NOT NULL,
	age INT,
	city_id INT,
	PRIMARY KEY (id),
	FOREIGN KEY (city_id)
		REFERENCES city(id)
		ON DELETE SET NULL
)
`

var interestSchema = `
CREATE TABLE IF NOT EXISTS interest (
	id INT NOT NULL AUTO_INCREMENT,
	name VARCHAR(64) NOT NULL UNIQUE,
	PRIMARY KEY (id)
)
`

var userInterestsSchema = `
CREATE TABLE IF NOT EXISTS user_interest (
	id INT NOT NULL AUTO_INCREMENT,
	user_id INT NOT NULL,
	interest_id INT NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (user_id)
		REFERENCES user(id)
		ON DELETE CASCADE,
	FOREIGN KEY (interest_id)
		REFERENCES interest(id)
		ON DELETE CASCADE
)
`

var sessionSchema = `
CREATE TABLE IF NOT EXISTS session (
	user_id INT NOT NULL,
	name VARCHAR(64) NOT NULL,
	value VARCHAR(255) NOT NULL,
	FOREIGN KEY (user_id)
		REFERENCES user(id)
		ON DELETE CASCADE
)
`
