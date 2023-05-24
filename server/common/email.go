package common

import (
	"fmt"
	"github.com/Leantar/elonwallet-backend/config"
	"net/smtp"
	"strings"
)

func SendEmail(cfg config.EmailConfig, recipient, title, body string) error {
	receiver := []string{recipient}
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("From: %s\r\n", cfg.User))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", recipient))
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n\r\n", title))
	builder.WriteString(fmt.Sprintf("%s\r\n", body))

	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.AuthHost)
	err := smtp.SendMail(cfg.SmtpHost, auth, cfg.User, receiver, []byte(builder.String()))
	if err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
	}
	return nil
}
