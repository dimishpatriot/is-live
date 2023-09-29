build:
	go build -o ./bin/app

run:
	go run main.go

run-short:
	go run main.go -s short -t 15s -f 5s -l 10

test:
	go test ./...