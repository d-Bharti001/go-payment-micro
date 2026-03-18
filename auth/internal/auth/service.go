package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

func (svc *AuthService) GetToken(ctx context.Context, credentials *pb.Credentials) (*pb.Token, error) {
	type user struct {
		userID   string
		password string
	}

	var u user

	stmt, err := svc.db.Prepare(`
		SELECT user_id, password
		FROM users
		WHERE user_id = ? AND password = ?
	`)
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = stmt.QueryRow(
		credentials.GetUserName(),
		credentials.GetPassword(),
	).Scan(
		&u.userID,
		&u.password,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	jwt, err := createJWT(u.userID)
	if err != nil {
		return nil, err
	}

	return &pb.Token{
		Jwt: jwt,
	}, nil
}

func (svc *AuthService) ValidateToken(ctx context.Context, token *pb.Token) (*pb.User, error) {
	userID, err := validateJWT(token.Jwt)
	if err != nil {
		return nil, err
	}

	return &pb.User{
		UserId: userID,
	}, nil
}
