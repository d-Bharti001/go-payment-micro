package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "github.com/d-Bharti001/go-payment-micro/auth/proto"
	mmpb "github.com/d-Bharti001/go-payment-micro/money_movement/proto"
)

var authClient authpb.AuthServiceClient
var mmClient mmpb.MoneyMovementServiceClient

func main() {

	// GRPC Connection to Auth service

	authConn, err := grpc.NewClient("auth:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := authConn.Close(); err != nil {
			log.Println(err)
		}
	}()

	authClient = authpb.NewAuthServiceClient(authConn)

	// GRPC Connection to Money Movement service

	mmConn, err := grpc.NewClient("money_movement:7000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := mmConn.Close(); err != nil {
			log.Println(err)
		}
	}()

	mmClient = mmpb.NewMoneyMovementServiceClient(mmConn)

	// HTTP Server

	http.HandleFunc("/login", login)
	http.HandleFunc("/customer/payment/authorize", customerPaymentAuthorize)
	http.HandleFunc("/customer/payment/capture", customerPaymentCapture)
}

func login(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token, err := authClient.GetToken(
		context.Background(),
		&authpb.Credentials{
			UserName: username,
			Password: password,
		},
	)

	if err != nil {
		_, writeErr := w.Write([]byte(err.Error()))
		if writeErr != nil {
			log.Println(writeErr)
		}
		return
	}

	_, writeErr := w.Write([]byte(token.Jwt))
	if writeErr != nil {
		log.Println(writeErr)
	}
}

func customerPaymentAuthorize(w http.ResponseWriter, r *http.Request) {
	// Get the JWT from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Extract the request payload
	type authorizePayload struct {
		MerchantUserId string `json:"merchant_user_id"`
		Paise          int64  `json:"paise"`
	}

	var payload authorizePayload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Validate JWT from auth service
	user, err := authClient.ValidateToken(
		context.Background(),
		&authpb.Token{
			Jwt: token,
		},
	)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Execute authorize payment transaction (money movement service)
	authorizeResponse, err := mmClient.Authorize(
		context.Background(),
		&mmpb.AuthorizePayload{
			CustomerUserId: user.GetUserId(),
			MerchantUserId: payload.MerchantUserId,
			Paise:          payload.Paise,
		},
	)
	if err != nil {
		_, writeErr := w.Write([]byte(err.Error()))
		if writeErr != nil {
			log.Println(writeErr)
		}
		return
	}

	// Return PID of authorize transaction
	type response struct {
		Pid string `json:"pid"`
	}

	resp := response{
		Pid: authorizeResponse.Pid,
	}

	respJson, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, writeErr := w.Write(respJson)
	if writeErr != nil {
		log.Println(writeErr)
	}
}

func customerPaymentCapture(w http.ResponseWriter, r *http.Request) {
	// Get the JWT from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Extract the request payload
	type capturePayload struct {
		Pid string `json:"pid"`
	}

	var payload capturePayload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Validate JWT from auth service
	user, err := authClient.ValidateToken(
		context.Background(),
		&authpb.Token{
			Jwt: token,
		},
	)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Execute capture payment transaction (money movement service)
	_, err = mmClient.Capture(
		context.Background(),
		&mmpb.CapturePayload{
			Pid:            payload.Pid,
			CustomerUserId: user.GetUserId(),
		},
	)
	if err != nil {
		_, writeErr := w.Write([]byte(err.Error()))
		if writeErr != nil {
			log.Println(writeErr)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
