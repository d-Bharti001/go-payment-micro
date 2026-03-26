package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/IBM/sarama"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"

	mm "github.com/d-Bharti001/go-payment-micro/internal/money_movement"
	pb "github.com/d-Bharti001/go-payment-micro/proto"
)

const (
	dbDriver = "mysql"
	dbName   = "money_movement"
)

var db *sql.DB
var msgProducer sarama.SyncProducer

func main() {
	var err error

	dbUser := os.Getenv("MYSQL_USERNAME")
	dbPassword := os.Getenv("MYSQL_PASSWORD")

	// Open a database connection
	dsn := fmt.Sprintf("%s:%s@tcp(mysql-money-movement:3306)/%s", dbUser, dbPassword, dbName)

	db, err = sql.Open(dbDriver, dsn)
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

	// Sarama Producer setup
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Return.Errors = true
	saramaCfg.Producer.Retry.Max = 5
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	sarama.Logger = log.New(os.Stdout, "[sarama]", log.LstdFlags)
	msgProducer, err = sarama.NewSyncProducer([]string{"my-cluster-kafka-bootstrap:9092"}, saramaCfg)
	if err != nil {
		log.Fatalf("Error creating message queue producer: %s", err)
	}

	defer func() {
		if err = msgProducer.Close(); err != nil {
			log.Printf("Error closing message queue producer: %s", err)
		}
	}()

	// GRPC Server setup
	grpcServer := grpc.NewServer()
	mmService := mm.NewMoneyMovementService(db, msgProducer)
	pb.RegisterMoneyMovementServiceServer(grpcServer, mmService)

	// Listen and serve
	listener, err := net.Listen("tcp", ":7000")
	if err != nil {
		log.Fatalf("failed to listed on port 7000: %v", err)
	}

	log.Printf("server listening at %v", listener.Addr())
	if err = grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}
