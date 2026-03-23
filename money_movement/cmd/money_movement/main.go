package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/IBM/sarama"
	"google.golang.org/grpc"

	mm "github.com/d-Bharti001/go-payment-micro/internal/money_movement"
	pb "github.com/d-Bharti001/go-payment-micro/proto"
)

const (
	dbDriver   = "mysql"
	dbUser     = "root"
	dbPassword = "Admin123"
	dbName     = "money_movement"
)

var db *sql.DB
var msgProducer sarama.SyncProducer

func main() {
	var err error

	// Open a database connection
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s", dbUser, dbPassword, dbName)

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
	msgProducer, err = sarama.NewSyncProducer([]string{"localhost:9092"}, sarama.NewConfig())
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
