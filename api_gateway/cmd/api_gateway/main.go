package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/d-Bharti001/go-payment-micro/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	authpb "github.com/d-Bharti001/go-payment-micro/auth/proto"
	mmpb "github.com/d-Bharti001/go-payment-micro/money_movement/proto"
)

var authClient authpb.AuthServiceClient
var mmClient mmpb.MoneyMovementServiceClient

// @title						Payment API Gateway
// @version					1.0
// @description				API Gateway service for payment app
// @BasePath					/
//
// @securityDefinitions.basic	BasicAuth
// @in							header
// @name						Authorization
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Enter the token with the `Bearer ` prefix, e.g. `Bearer <jwt>`
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

	mmConn, err := grpc.NewClient("money-movement:7000", grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	// Swagger docs
	http.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	http.HandleFunc("GET /login", login)
	http.HandleFunc("POST /customer/payment/authorize", customerPaymentAuthorize)
	http.HandleFunc("POST /customer/payment/capture", customerPaymentCapture)

	fmt.Println("Listening on port 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Login godoc
//
//	@Summary		Customer login
//	@Description	Login endpoint for a customer. Required before initiating a payment.
//
//	@Tags			login
//	@Produce		plain
//	@Security		BasicAuth
//
//	@Param			Authorization	header		string	true	"Base64 encoded username and password, e.g., 'Basic base64(username:password)'"
//
//	@Success		200				{string}	string	"JWT token"
//	@Failure		401				{string}	string
//	@Router			/login [get]
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
		w.WriteHeader(http.StatusUnauthorized)
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

type AuthorizePayload struct {
	MerchantUserId string `json:"merchant_user_id"`
	Paise          int64  `json:"paise"`
}

type AuthorizeResponse struct {
	Pid string `json:"pid"`
}

// CustomerPaymentAuthorize godoc
//
//	@Summary		Authorize a customer payment
//	@Description	Authorizes a payment from the authenticated customer to the merchant.
//	@Description	Transfers the payment amount from the customer's Default account to their Payment account.
//	@Description	Returns a PID used to capture the payment as the next step.
//	@Tags			payment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			payload	body		AuthorizePayload	true	"Authorize payment payload"
//	@Success		200		{object}	AuthorizeResponse
//	@Failure		400		{string}	string
//	@Failure		401		{string}	string
//	@Failure		500		{string}	string
//	@Router			/customer/payment/authorize [post]
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
	var payload AuthorizePayload
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return PID of authorize transaction
	resp := AuthorizeResponse{
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

type CapturePayload struct {
	Pid string `json:"pid"`
}

// CustomerPaymentCapture godoc
//
//	@Summary		Capture a payment
//	@Description	Captures (finalizes) an authorized payment using its PID.
//	@Description	Sends the payment amount from the customer's Payment account to the merchant's Incoming account.
//	@Tags			payment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			payload	body		CapturePayload	true	"Capture payload"
//	@Success		200		{string}	string			"OK (empty body)"
//	@Failure		400		{string}	string
//	@Failure		401		{string}	string
//	@Failure		500		{string}	string
//	@Router			/customer/payment/capture [post]
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
	var payload CapturePayload
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
