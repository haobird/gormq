package main

import (
	"fmt"
	"time"

	"github.com/haobird/gormq"
	"github.com/streadway/amqp"
)

func main() {
	extras := []gormq.ExtraFunc{
		gormq.SetQueueDurable(false),
		gormq.SetQueueAutoDelete(false),
		gormq.SetQueueArgs(map[string]interface{}{
			"x-message-ttl": 30000,
		}),
	}
	queueExchange := gormq.QueueExchange{
		QueueName:    "myname11",
		RoutingKey:   "iot.report.opendoor.insert",
		ExchangeName: "iot.opendoor",
		ExchangeType: "topic",
	}
	addr := "amqp://admin:admin@127.0.0.1:5672/"
	rabbitmq := gormq.New(addr)

	publisher := rabbitmq.NewPublisher(queueExchange, extras...)
	go func() {
		for {
			err := publisher.Pub([]byte("hhshs"))
			fmt.Println(err)
			time.Sleep(2 * time.Second)
		}
	}()

	consumer := rabbitmq.NewConsumer(queueExchange, extras...)
	// # 消费消息
	consumer.Receive(func(msg amqp.Delivery) {
		msg.Ack(false)
		fmt.Println(string(msg.Body))
	})
}
