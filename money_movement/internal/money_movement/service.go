package mm

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/d-Bharti001/go-payment-micro/internal/producer"
	pb "github.com/d-Bharti001/go-payment-micro/proto"
)

type MoneyMovementService struct {
	db *sql.DB
	pb.UnimplementedMoneyMovementServiceServer
}

func NewMoneyMovementService(db *sql.DB) *MoneyMovementService {
	return &MoneyMovementService{
		db: db,
	}
}

func (svc *MoneyMovementService) Authorize(ctx context.Context, payload *pb.AuthorizePayload) (*pb.AuthorizeResponse, error) {
	// Begin db transaction
	tx, err := svc.db.Begin()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	merchantWallet, err := fetchWalletByUserID(tx, payload.GetMerchantWalletUserId(), CUSTOMER_WALLET)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	customerWallet, err := fetchWalletByUserID(tx, payload.GetCustomerWalletUserId(), MERCHANT_WALLET)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	srcAccount, err := fetchAccountByWalletID(tx, customerWallet.id, ACC_DEFAULT)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	dstAccount, err := fetchAccountByWalletID(tx, customerWallet.id, ACC_PAYMENT)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = transfer(tx, srcAccount, dstAccount, payload.GetPaise())
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	pid := uuid.NewString()

	err = createTransaction(tx, pid, srcAccount, dstAccount, customerWallet, customerWallet, merchantWallet, payload.GetPaise())
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit the db transaction
	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AuthorizeResponse{
		Pid: pid,
	}, nil
}

func (svc *MoneyMovementService) Capture(ctx context.Context, payload *pb.CapturePayload) (*emptypb.Empty, error) {
	// Begin db transaction
	tx, err := svc.db.Begin()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	authorizeTransaction, err := fetchTransactionForCapture(tx, payload.GetPid())
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	srcAccount, err := fetchAccountByID(tx, authorizeTransaction.dstAccountId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	srcWallet, err := fetchWalletByID(tx, authorizeTransaction.dstWalletId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	dstMerchantAccount, err := fetchAccountByWalletID(tx, authorizeTransaction.finalDstMerchantWalletId, ACC_INCOMING)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	dstMerchantWallet, err := fetchWalletByID(tx, authorizeTransaction.finalDstMerchantWalletId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = transfer(tx, srcAccount, dstMerchantAccount, authorizeTransaction.amount)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = createTransaction(tx, authorizeTransaction.pid, srcAccount, dstMerchantAccount, srcWallet, dstMerchantWallet, dstMerchantWallet, authorizeTransaction.amount)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit the db transaction
	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
