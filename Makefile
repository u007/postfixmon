build:
	go build -o bin/postfixmon main.go

run:
	API_TOKEN=aaa go run main.go reset
	API_TOKEN=aaa go run main.go run

rerun:
	# make rerun date="2022 oct 16"
	API_TOKEN=aaa go run main.go rerun $(date)