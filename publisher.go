package main

import (
	"github.com/streadway/amqp"
	"os"
	"log"
	"fmt"
	"strings"
	"io/ioutil"
	"math/rand"
	"time"
)

type DebtwireAsLive struct {
}

// sends an intel to rabbit
func (d *DebtwireAsLive) Publish(someBody string, title string) error {
	fmt.Println("INTEL_STORE_RABBIT_URL:", os.Getenv("INTEL_STORE_RABBIT_URL"))
	fmt.Println("INTEL_EXCHANGE_NAME:", os.Getenv("INTEL_EXCHANGE_NAME"))
	fmt.Println("TEST:", os.Getenv("TEST_EXCHANGE"))
	rabbitURL := os.Getenv("INTEL_STORE_RABBIT_URL")
	rabbitExchange := os.Getenv("INTEL_EXCHANGE_NAME")
	testExchange := os.Getenv("TEST_EXCHANGE")
	conn, err := amqp.Dial(rabbitURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		rabbitExchange,   // name
		"topic", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	body, err := ioutil.ReadFile("./intel.json")

	id := RandomString(10)

	replacedBody := fmt.Sprintf(string(body), id, someBody)

	err = ch.Publish(
		testExchange, // exchange
		"PublishIntelligence.V2",     // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(replacedBody),
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent something")
	return nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}


func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
