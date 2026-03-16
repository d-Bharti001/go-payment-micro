package auth

import (
	"context"
	"database/sql"

	pb "github.com/d-Bharti001/go-payment-micro/proto"
)

type AuthService struct {
	db *sql.DB
	pb.UnimplementedAuthServiceServer
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{
		db: db,
	}
}

// TODO
func (svc *AuthService) GetToken(ctx context.Context, credentials *pb.Credentials) (*pb.Token, error) {
	type user struct {
		userID   string
		password string
	}

	var u user

	stmt, err := svc.db.Prepare()

	return &pb.Token{}, nil
}

// TODO
func (svc *AuthService) ValidateToken(ctx context.Context, token *pb.Token) (*pb.User, error) {
	return &pb.User{}, nil
}
