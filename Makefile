
generate: mocks


mocks:
	echo "Generating mocks"

migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)