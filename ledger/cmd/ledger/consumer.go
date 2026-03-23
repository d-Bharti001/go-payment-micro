package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"

	"github.com/d-Bharti001/go-payment-micro/internal/ledger"
)

const (
	dbDriver   = "mysql"
	dbUser     = "root"
	dbPassword = "Admin123"
	dbName     = "ledger"
	topic      = "ledger"
)

var (
	db *sql.DB
	wg sync.WaitGroup
)

type LedgerMsg struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	Amount    int64  `json:"amount"`
	Operation string `json:"operation"`
	Time      string `json:"time"`
}

func main() {
	// Open a database connection
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s", dbUser, dbPassword, dbName)

	db, err := sql.Open(dbDriver, dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = db.Close(); err != nil {
			log.Printf("Error closing db: %s", err)
		}
	}()

	// Ping the database to ensure connection is valid
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})

	// Create a message queue consumer
	consumer, err := sarama.NewConsumer([]string{"kafka:9092"}, sarama.NewConfig())
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		close(done)
		if err := consumer.Close(); err != nil {
			log.Println(err)
		}
	}()

	// Get all the partitions of the concerned topic
	partitions, err := consumer.Partitions(topic)
	if err != nil {
		log.Fatal(err)
	}

	// Consume messages from all the partitions
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
	var ledgerMsg LedgerMsg

	err := json.Unmarshal(msg.Value, &ledgerMsg)
	if err != nil {
		fmt.Println("Cannot unmarshal message:", err)
		return
	}

	err = ledger.Insert(
		db,
		ledgerMsg.OrderID,
		ledgerMsg.UserID,
		ledgerMsg.Amount,
		ledgerMsg.Operation,
		ledgerMsg.Time,
	)
	if err != nil {
		fmt.Println("Cannot send email:", err)
		return
	}
}
