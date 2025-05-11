all:
	go build -o ./handlers_gen.exe handlers_gen/*
	./handlers_gen.exe api.go api_handlers.go

gen:
	go build handlers_gen/* && ./codegen api.go api_handlers.go

test:
	go test -v