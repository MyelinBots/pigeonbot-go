
generate: mocks


mocks:
	go get go.uber.org/mock/mockgen/model
	go install go.uber.org/mock/mockgen@latest
	go generate ./...


migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)