run:
	go run main.go

unit-tests:
	go test ./...

build:
	docker build . -t whobson00/homieclips-hls-converter:latest