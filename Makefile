build:
	# go build -o bin/postfixmon main.go
	env GOOS=linux GOARCH=amd64 go build -o bin/


run:
	API_TOKEN=aaa go run main.go reset
	API_TOKEN=aaa PF_LOG=mail.log go run main.go run

rerun:
	# make rerun date="2022 oct 16"
	API_TOKEN=aaa go run main.go rerun $(date)