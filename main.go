package main

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// Publisher with reconnect logic
func publishForever() {
	for {
		err := connectAndPublish()
		if err != nil {
			log.Printf("Publisher: encountered error: %v. Reconnecting in 5s...", err)
			time.Sleep(5 * time.Second)
		}
		log.Printf("error was nill shit oh my god")
	}
}

func connectAndPublish() error {
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"my_exchange", // name
		"direct",      // type
		true,          // durable
		false,         // autoDelete
		false,         // internal
		false,         // noWait
		nil,           // args
	)

	failOnError(err, "Failed to declare exchange")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for t := range ticker.C {
		body := "Hello at " + t.Format(time.RFC3339)
		err = ch.Publish("my_exchange", "my_key", false, false, amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})

		if err != nil {
			log.Printf("Publisher: publish error: %v", err)
			return fmt.Errorf("publish failed, will reconnect: %w", err)
		} else {
			log.Printf("Publisher: sent -> %s", body)
		}
	}
	return nil
}

func connectAndConsume() error {
	conn, err := amqp091.DialConfig("amqp://guest:guest@localhost:5672/", amqp091.Config{
		Heartbeat: 10 * time.Second, // Send heartbeat every 10s
		Locale:    "en_US",
	})
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("my_queue", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind("my_queue", "my_key", "my_exchange", false, nil)
	if err != nil {
		return fmt.Errorf("faield to bind exhchange to queue.perhaps the producer does not came up ..... wait for producer to up")
	}

	msgs, err := ch.Consume(
		q.Name, // queue name
		"",     // consumer tag
		false,  // autoAck: set to false for manual acknowledgment
		false,  // exclusive
		false,  // noLocal
		false,  // noWait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Listen for channel or connection close
	closeChan := make(chan *amqp091.Error)
	ch.NotifyClose(closeChan)

	log.Println("Consumer: waiting for messages")

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed unexpectedly")
			}
			log.Printf("Consumer: received -> %s", msg.Body)

			// Simulate message processing
			time.Sleep(30 * time.Second)
			ackDone := msg.Ack(true)
			if ackDone != nil {
				return fmt.Errorf("faield too acknoolege")
			}

		case err := <-closeChan:
			if err != nil {
				return fmt.Errorf("connection/channel closed: %v", err)
			}
			return fmt.Errorf("connection/channel closed normally")
		}
	}

}

func consumeForever() {

	for {
		err := connectAndConsume()
		if err != nil {
			log.Printf("Consumer error: %v. Reconnecting in 5s...", err)
			time.Sleep(5 * time.Second)
		}
		log.Printf("error was nill shit oh my god")
	}

}
func main() {
	go publishForever()
	go consumeForever()

	// Block main thread forever
	select {}
}
