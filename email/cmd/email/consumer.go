package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/IBM/sarama"

	"github.com/d-Bharti001/go-payment-micro/internal/email"
)

const topic = "email"

var wg sync.WaitGroup

type EmailMsg struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
}

func main() {
	done := make(chan struct{})

	sarama.Logger = log.New(os.Stdout, "[sarama]", log.LstdFlags)
	consumer, err := sarama.NewConsumer([]string{"my-cluster-kafka-bootstrap:9092"}, sarama.NewConfig())
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		close(done)
		if err := consumer.Close(); err != nil {
			log.Println(err)
		}
	}()

	partitions, err := consumer.Partitions(topic)
	if err != nil {
		log.Fatal(err)
	}

	for _, partition := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			if err := partitionConsumer.Close(); err != nil {
				log.Println("Partition consumer close error", err)
			}
		}()

		wg.Go(func() { awaitMessages(done, partitionConsumer, partition) })
	}

	wg.Wait()
}

func awaitMessages(done chan struct{}, partitionConsumer sarama.PartitionConsumer, partition int32) {
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			fmt.Printf("Partition: %d - Received message: %s\n", partition, string(msg.Value))
			handleMessage(msg)

		case <-done:
			fmt.Println("Received done signal, exiting...")
			return
		}
	}
}

func handleMessage(msg *sarama.ConsumerMessage) {
	var emailMsg EmailMsg

	err := json.Unmarshal(msg.Value, &emailMsg)
	if err != nil {
		fmt.Println("Cannot unmarshal message:", err)
		return
	}

	err = email.Send(emailMsg.UserID, emailMsg.OrderID)
	if err != nil {
		fmt.Println("Cannot send email:", err)
		return
	}
}
