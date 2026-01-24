USER=root
SERVER=example.com
include .env

build:
	# go build -o bin/postfixmon main.go
	env GOOS=linux GOARCH=amd64 go build -o bin/
	env GOOS=linux GOARCH=amd64 go build -o bin/pffind postfix/main/main.go 

run:
	API_TOKEN=aaa go run main.go reset
	API_TOKEN=aaa PF_LOG=mail.log go run main.go run

rerun:
	# make rerun date="2022 oct 16"
	API_TOKEN=aaa go run main.go rerun $(date)

upload:
	make build
	ssh -t $(USER)@$(SERVER) "pkill postfixmon" & true
	scp bin/postfixmon $(USER)@$(SERVER):/root/postfixmon/
	scp bin/pffind $(USER)@$(SERVER):/root/postfixmon/

testdeploy: upload
	ssh -t $(USER)@$(SERVER) "cd postfixmon && bash once.sh"

deploy: upload
	ssh -t $(USER)@$(SERVER) "cd postfixmon && bash run.sh"

ssh:
	ssh -t $(USER)@$(SERVER) "sh -c 'cd postfixmon && bash'"

logs:
	ssh -t $(USER)@$(SERVER) "cd postfixmon && tail -n 200 -f out.log"

testfind: build
	# scp bin/pffind $(USER)@$(SERVER):/root/postfixmon/
	# ssh -t $(USER)@$(SERVER) "sh -c 'cd postfixmon && ./pffind fais@msnwork.com.my'"
	# ./bin/pffind mail.log
	go run postfix/main/main.go azira@mnswork.com.my
	
