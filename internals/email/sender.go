package email

import (
	"email-blaze/internals/config"
	"fmt"
	"net/smtp"
)

type Sender struct {
	config *config.Config
}

func NewSender(cfg *config.Config) *Sender {
	return &Sender{config: cfg}
}

func (s *Sender) Send(from, to, subject, body string) error {
	msg := []byte("To: " + to + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body)

	auth := smtp.PlainAuth("", from, s.config.SMTPPassword, s.config.SMTPHost)
	return smtp.SendMail(fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort), auth, from, []string{to}, msg)
}