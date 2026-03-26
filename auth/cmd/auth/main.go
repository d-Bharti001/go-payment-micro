package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"

	"github.com/d-Bharti001/go-payment-micro/internal/auth"
	pb "github.com/d-Bharti001/go-payment-micro/proto"
)

const (
	dbDriver = "mysql"
	dbName   = "auth"
)

var db *sql.DB

func main() {
	var err error

	dbUser := os.Getenv("MYSQL_USERNAME")
	dbPassword := os.Getenv("MYSQL_PASSWORD")

	// Open a database connection
	dsn := fmt.Sprintf("%s:%s@tcp(mysql-auth:3306)/%s", dbUser, dbPassword, dbName)

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

	// GRPC Server setup
	grpcServer := grpc.NewServer()
	authService := auth.NewAuthService(db)
	pb.RegisterAuthServiceServer(grpcServer, authService)

	// Listen and serve
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen on port 9000: %v", err)
	}

	log.Printf("server listening at %v", listener.Addr())
	if err = grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
