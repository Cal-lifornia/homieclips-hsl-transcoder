package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func setupRabbitMQ(conn *amqp.Connection) *amqp.Channel {

	rabbitCh, err := conn.Channel()
	failOnError(err, "failed to open a channel")
	queue, err := rabbitCh.QueueDeclare(
		"uploaded_files",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "failed to declare queue")

	err = rabbitCh.QueueBind(
		queue.Name,
		"uploaded_files",
		queue.Name,
		false,
		nil,
	)

	failOnError(err, "failed to bind channel")

	return rabbitCh
}
