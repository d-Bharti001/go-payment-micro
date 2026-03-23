package producer

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

const (
	emailTopic  = "email"
	ledgerTopic = "ledger"
)

type EmailMsg struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
}

type LedgerMsg struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	Amount    int64  `json:"amount"`
	Operation string `json:"operation"`
	Time      string `json:"time"`
}

func SendCaptureMessage(msgProducer sarama.SyncProducer, pid string, userID string, amount int64) {
	emailMsg := EmailMsg{
		OrderID: pid,
		UserID:  userID,
	}

	ledgerMsg := LedgerMsg{
		OrderID:   pid,
		UserID:    userID,
		Amount:    amount,
		Operation: "DEBIT",
		Time:      time.Now().Format("2006-01-02 15:04"),
	}

	var wg sync.WaitGroup

	wg.Go(func() { sendMessage(msgProducer, emailMsg, emailTopic) })
	wg.Go(func() { sendMessage(msgProducer, ledgerMsg, ledgerTopic) })

	wg.Wait()
}

func sendMessage[T EmailMsg | LedgerMsg](msgProducer sarama.SyncProducer, msg T, topic string) {
	strMsg, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshalling message: %s\n", err)
		return
	}

	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(strMsg),
	}

	partition, offset, err := msgProducer.SendMessage(message)
	if err != nil {
		log.Printf("Error sending message to queue: %s\n", err)
		return
	}

	log.Printf("Message send to partition %d at offset %d\n", partition, offset)
}
