# Pigeon Bot go
A irc bot written in go

## how to run
1. install modules
    `$ go mod tidy`
2. start docker for postgres
    `$ docker compose up -d`
3. Migrate up
    `$ go run cmd/main.go migrate up` or `$ make migrate-up`
3. start the bot
    `$ go run cmd/main.go serve`
