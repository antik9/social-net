package projecterrors

import (
	"log"
)

const (
	UnknownDatabaseError = "Unknown error on server, please try again"
	UnknownTemplateError = "Unknown error, please try again"
	UserAlreadyExists    = "User with this email already exists"
)

func FailOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
