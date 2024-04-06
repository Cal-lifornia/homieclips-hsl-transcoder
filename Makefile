run:
	go run main.go

unit-tests:
	go test ./...

build:
	docker build . -t homieclips/hls-converter