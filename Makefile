.PHONY: all

all: migrate server

migrate:
	go build -o sn-migrate cmd/migrate/main.go

server:
	go build -o sn-server cmd/server/main.go

queue:
	go build -o sn-queue cmd/queue/main.go
