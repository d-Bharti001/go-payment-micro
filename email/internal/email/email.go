package email

import (
	"fmt"
	"net/smtp"
)

func Send(target string, orderID string) error {

	// TODO
	senderEmail := "my_email@gmail.com"
	password := "my_password"

	recipientEmail := target

	message := []byte(fmt.Sprintf("Subject: Payment Processed!\n Process ID: %s\n", orderID))

	smtpServer := "smtp.gmail.com"
	smtpPort := 587

	creds := smtp.PlainAuth("", senderEmail, password, smtpServer)

	smtpAddress := fmt.Sprintf("%s:%d", smtpServer, smtpPort)

	return smtp.SendMail(smtpAddress, creds, senderEmail, []string{recipientEmail}, message)
}
