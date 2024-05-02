FROM golang:1.21-bullseye as build-stage
LABEL authors="whobson"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./hls-converter .

FROM alpine:3.19 as build-release-stage
WORKDIR /

COPY --from=build-stage /app/hls-converter /hls-converter

RUN touch .env

RUN apk update
RUN apk upgrade
RUN apk add --no-cache ffmpeg

ENV ENVIRONMENT="docker"

ENTRYPOINT ["/hls-converter"]