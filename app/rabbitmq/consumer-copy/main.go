package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	cfg "github.com/pobyzaarif/go-config"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	RabbitMQURL string `env:"RABBITMQ_URL"`
}

type ProductMessage struct {
	ProductCode string `json:"product_code"`
	ProductName string `json:"product_name"`
	Stock       int    `json:"stock"`
}

func main() {
	config := Config{}
	err := cfg.LoadConfig(&config)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	queue := "product-queue"

	conn, err := amqp.Dial(config.RabbitMQURL)
	if err != nil {
		log.Fatalf("dial err: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel err: %v", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("queue declare err: %v", err)
	}

	msgs, err := ch.Consume(
		queue,
		"",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("consume err: %v", err)
	}

	// seed rng for simulated failures
	randSource := rand.New(rand.NewSource(time.Now().UnixNano()))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}

				var p ProductMessage
				if err := json.Unmarshal(d.Body, &p); err != nil {
					log.Println("JSON decode error:", err)
					_ = d.Nack(false, false)
					continue
				}

				log.Printf("Received Product:\n- Code: %s\n- Name: %s\n- Stock: %d\n",
					p.ProductCode, p.ProductName, p.Stock)

				// simulate processing
				time.Sleep(200 * time.Millisecond)

				// simulate failure sometimes so message will be requeued and
				// can be consumed by other consumers
				if randSource.Intn(2) == 0 { // ~50% chance to "fail"
					log.Printf("Simulated processing error for %s, Nack with requeue", p.ProductCode)
					_ = d.Nack(false, true) // requeue = true
					continue
				}

				_ = d.Ack(false)
			}
		}
	}()

	// wait for CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("consumer shutting down")
}
