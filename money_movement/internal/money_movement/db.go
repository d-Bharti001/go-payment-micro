package mm

import (
	"database/sql"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func fetchWalletByUserID(tx *sql.Tx, userID string, walletType walletType) (*wallet, error) {
	stmt, err := tx.Prepare(`
		SELECT id, user_id, wallet_type
		FROM wallets
		WHERE
			user_id = ?
			AND wallet_type = ?
	`)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var w wallet
	err = stmt.QueryRow(
		userID,
		walletType,
	).Scan(
		&w.id,
		&w.userId,
		&w.walletType,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &w, nil
}

func fetchWalletByID(tx *sql.Tx, walletID int32) (*wallet, error) {
	stmt, err := tx.Prepare(`
		SELECT id, user_id, wallet_type
		FROM wallets
		WHERE id = ?
	`)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var w wallet
	err = stmt.QueryRow(walletID).Scan(
		&w.id,
		&w.userId,
		&w.walletType,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &w, nil
}

func fetchAccountByWalletID(tx *sql.Tx, walletID int32, accType accountType) (*account, error) {
	stmt, err := tx.Prepare(`
		SELECT id, paise, account_type, wallet_id
		FROM accounts
		WHERE
			wallet_id = ?
			AND account_type = ?
	`)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var a account
	err = stmt.QueryRow(
		walletID,
		accType,
	).Scan(
		&a.id,
		&a.paise,
		&a.accountType,
		&a.walletId,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &a, nil
}

func fetchAccountByID(tx *sql.Tx, accountID int32) (*account, error) {
	stmt, err := tx.Prepare(`
		SELECT id, paise, account_type, wallet_id
		FROM accounts
		WHERE id = ?
	`)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var a account
	err = stmt.QueryRow(accountID).Scan(
		&a.id,
		&a.paise,
		&a.accountType,
		&a.walletId,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &a, nil
}

func transfer(tx *sql.Tx, srcAccount *account, dstAccount *account, amount int64) error {
	// Check for enough amount
	if srcAccount.paise < amount {
		return status.Error(codes.InvalidArgument, "not enough money")
	}

	// Subtract the amount from source account
	_, err := tx.Exec(
		`
			UPDATE accounts
			SET paise = paise - ?
			WHERE id = ?
		`,
		amount,
		srcAccount.id,
	)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// Add the amount to the destination account
	_, err = tx.Exec(
		`
			UPDATE accounts
			SET paise = paise + ?
			WHERE id = ?
		`,
		amount,
		dstAccount.id,
	)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func createTransaction(
	tx *sql.Tx,
	pid string,
	srcAccount, dstAccount *account,
	srcWallet, dstWallet, merchantWallet *wallet,
	amount int64,
) error {
	stmt, err := tx.Prepare(`
		INSERT INTO transactions
		(
			pid,
			src_user_id, dst_user_id,
			src_wallet_id, dst_wallet_id,
			src_account_id, dst_account_id,
			src_account_type, dst_account_type,
			final_dst_merchant_wallet_id,
			amount
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	_, err = stmt.Exec(
		pid,
		srcWallet.userId, dstWallet.userId,
		srcAccount.walletId, dstAccount.walletId,
		srcAccount.id, dstAccount.id,
		srcAccount.accountType, dstAccount.accountType,
		merchantWallet.id,
		amount,
	)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func fetchTransactionForCapture(tx *sql.Tx, pid string, customerUserID string) (*transaction, error) {
	stmt, err := tx.Prepare(`
		SELECT
			id,
			pid,
			src_user_id,
			dst_user_id,
			src_wallet_id,
			dst_wallet_id,
			src_account_id,
			dst_account_id,
			src_account_type,
			dst_account_type,
			final_dst_merchant_wallet_id,
			amount
		FROM transactions
		WHERE
			pid = ? AND src_user_id = ?
	`)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rows, err := stmt.Query(pid, customerUserID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	var result *transaction

	for rows.Next() {
		var t transaction
		err = rows.Scan(
			&t.id,
			&t.pid,
			&t.srcUserId,
			&t.dstUserId,
			&t.srcWalletId,
			&t.dstWalletId,
			&t.srcAccountId,
			&t.dstAccountId,
			&t.srcAccountType,
			&t.dstAccountType,
			&t.finalDstMerchantWalletId,
			&t.amount,
		)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Transaction must be an AUTHORIZE transaction
		if !(t.srcAccountType == ACC_DEFAULT &&
			t.dstAccountType == ACC_PAYMENT) {
			return nil, status.Error(codes.Aborted, "transaction not authorized or already captured")
		}

		if result != nil {
			return nil, status.Error(codes.Aborted, "transaction might have already been captured")
		}

		result = &t
	}

	if result == nil {
		return nil, status.Error(codes.InvalidArgument, "transaction not authorized")
	}

	return result, nil
}
