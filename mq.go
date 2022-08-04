package main

import (
	"github.com/Shopify/sarama"
	"log"
	"os"
	"time"
)

const MQ_TOPIC_SLOW_REQUEST = "zg_slow_request"

func SendMq(topic string, value string) {
	MQ_SERVERS := os.Getenv("KAFKA_SERVERS")

	config := sarama.NewConfig()
	// config.Producer.RequiredAcks = sarama.WaitForAll
	// config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 10 * time.Second
	p, err := sarama.NewSyncProducer([]string{MQ_SERVERS}, config)
	if err != nil {
		log.Println("SendMq init error:", err)
		return
	}
	defer p.Close()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}

	if _, _, err := p.SendMessage(msg); err != nil {
		log.Println("SendMq SendMessage error:", err)
		return
	}
}
