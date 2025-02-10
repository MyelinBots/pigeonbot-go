
generate: mocks


mocks:
	go get go.uber.org/mock/mockgen/model
	go install go.uber.org/mock/mockgen@latest
	go generate ./...


migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)

migrate-up:
	go run cmd/main.go migrate up

migrate-down:
	go run cmd/main.go migrate down