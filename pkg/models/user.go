package models

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"

	"github.com/antik9/social-net/internal/db"
	"github.com/antik9/social-net/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	City         *City      `json:"-"`
	CityId       int        `db:"city_id" json:"city_id"`
	Id           int        `db:"id" json:"id"`
	FirstName    string     `db:"first_name" json:"first_name"`
	LastName     string     `db:"last_name" json:"last_name"`
	Email        string     `db:"email" json:"email"`
	Age          int        `db:"age" json:"age"`
	PasswordHash string     `db:"password" json:"-"`
	Interests    []Interest `json:"-"`
}

type Interest struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type Session struct {
	user   *User
	UserId int    `db:"user_id"`
	Name   string `db:"name"`
	Value  string `db:"value"`
}

func passwordHash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func getUserByEmail(email string, withForeignKeys bool) *User {
	var users []User
	db.Db.Select(&users, "SELECT * FROM user WHERE email = ?", email)

	if len(users) == 0 {
		return nil
	}
	user := users[0]
	if withForeignKeys {
		var cities []City
		db.Db.Select(&cities, "SELECT * FROM city WHERE id = ?", user.CityId)
		if len(cities) == 1 {
			user.City = &cities[0]
		}

		db.Db.Select(
			&user.Interests,
			`SELECT * FROM interest WHERE id IN
			(SELECT interest_id FROM user_interest WHERE user_id = ?)`, user.Id,
		)
	}
	return &user
}

func GetUsersLimitBy(limit int) []User {
	var users []User
	db.Db.Select(
		&users,
		"SELECT id, first_name, last_name FROM user LIMIT ?",
		limit,
	)
	return users
}

func GetUsersByNamePrefix(prefix string, limit int) []User {
	var users []User
	prefix += "%"

	replica := db.GetRandomReplica()
	result := replica.Select(
		&users,
		`SELECT id, first_name, last_name FROM user WHERE first_name LIKE ? OR last_name LIKE ?
		ORDER BY id LIMIT ?`,
		prefix, prefix, limit,
	)
	if result != nil {
		panic(result)
	}
	return users
}

func NewSession(email, password string) *Session {
	if user := getUserByEmail(email, false); user != nil {
		err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err == nil {
			value := sha256.New()
			userId := strconv.Itoa(user.Id)
			value.Write([]byte(userId))
			stringValue := hex.EncodeToString(value.Sum(nil))
			db.Db.MustExec(
				"INSERT INTO session (user_id, name, value) VALUES(?, ?, ?)",
				user.Id, "sn-session", stringValue,
			)
			var sessions []Session
			db.Db.Select(
				&sessions, "SELECT * FROM session WHERE value = ? AND user_id = ?",
				stringValue, user.Id,
			)
			if len(sessions) > 0 {
				return &sessions[0]
			}
		}
	}
	return nil
}

func NewUser(
	cityName, firstName, lastName, email, password string,
	age int,
	interests []string,
) (*User, error) {
	if user := getUserByEmail(email, false); user != nil {
		return nil, errors.New(projecterrors.UserAlreadyExists)
	} else {
		city := getOrCreateCity(cityName)
		var _interests []Interest
		for _, interest := range interests {
			_interests = append(_interests, getOrCreateInterest(interest))
		}
		hash := passwordHash(password)
		db.Db.MustExec(
			`INSERT INTO user (first_name, last_name, email, password, age, city_id)
			VALUES (?, ?, ?, ?, ?, ?)`,
			firstName, lastName, email, hash, age, city.Id,
		)
		user := getUserByEmail(email, false)
		for _, interest := range _interests {
			user.AddInterest(interest)
		}
		user.ChangeCity(city)
		return user, nil
	}
}

func (u *User) GetCityName() string {
	return u.City.Name
}

func (u *User) GetInterests() []Interest {
	return u.Interests
}

func GetUserBySessionValue(value string) *User {
	var users []User
	db.Db.Select(
		&users,
		`SELECT user.* FROM user JOIN session ON user.id = session.user_id
		WHERE session.value = ?`,
		value,
	)
	if len(users) > 0 {
		return getUserByEmail(users[0].Email, true)
	}
	return nil
}

func GetUserById(id int) *User {
	var users []User
	db.Db.Select(
		&users,
		`SELECT * FROM user WHERE id = ?`,
		id,
	)
	if len(users) > 0 {
		return getUserByEmail(users[0].Email, true)
	}
	return nil
}

func (u *User) AddInterest(interest Interest) {
	for _, _interest := range u.Interests {
		if _interest.Name == interest.Name {
			return
		}
	}
	db.Db.MustExec(
		"INSERT INTO user_interest (user_id, interest_id) VALUES (?, ?)",
		u.Id, interest.Id,
	)
	u.Interests = append(u.Interests, interest)
}

func (u *User) ChangeCity(city City) {
	u.City = &city
	u.CityId = city.Id
	db.Db.MustExec("UPDATE user SET city_id = ? WHERE email = ?", u.CityId, u.Email)
}

func getOrCreateCity(cityName string) City {
	var cities []City
	db.Db.Select(&cities, "SELECT * FROM city WHERE name = ?", cityName)
	if len(cities) == 0 {
		db.Db.MustExec("INSERT INTO city (name) VALUES (?)", cityName)
		db.Db.Select(&cities, "SELECT * FROM city WHERE name = ?", cityName)
	}
	return cities[0]
}

func getOrCreateInterest(interestName string) Interest {
	var interests []Interest
	db.Db.Select(&interests, "SELECT * FROM interest WHERE name = ?", interestName)
	if len(interests) == 0 {
		db.Db.MustExec("INSERT INTO interest (name) VALUES (?)", interestName)
		db.Db.Select(&interests, "SELECT * FROM interest WHERE name = ?", interestName)
	}
	return interests[0]
}
