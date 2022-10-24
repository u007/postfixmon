build:
	go build -o bin/postfixmon main.go

run:
	API_TOKEN=aaa go run main.go reset
	API_TOKEN=aaa go run main.go run
