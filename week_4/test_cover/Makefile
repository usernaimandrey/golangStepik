test:
	go test -v -cover

cover:
	go test -v -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html

cover-prof:
	go test -coverprofile=cover.out

open-cover:
	go tool cover -html=cover.out -o cover.html
