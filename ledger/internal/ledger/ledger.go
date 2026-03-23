package ledger

import "database/sql"

func Insert(
	db *sql.DB,
	orderID string,
	userID string,
	amount int64,
	operation string,
	transactionTime string,
) error {

	stmt, err := db.Prepare(`
		INSERT INTO ledger (order_id, user_id, amount, operation, transaction_time)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(orderID, userID, amount, operation, transactionTime)

	return err
}
