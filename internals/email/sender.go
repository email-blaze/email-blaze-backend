package email

import (
	"bytes"
	"email-blaze/internals/config"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/mail"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

type Email struct {
	From    string
	To      []string
	Subject string
	Body    string
}

type Sender struct {
	config *config.Config
}

func NewSender(cfg *config.Config) *Sender {
	return &Sender{config: cfg}
}

func (s *Sender) Send(from, to, subject, body string, html bool, domain string) error {
	auth := sasl.NewPlainClient("", s.config.DefaultUser.Email, s.config.DefaultUser.Password)

	client, err := smtp.DialTLS(fmt.Sprintf("%s:%d", domain, s.config.SMTPPort), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	if err := client.Mail(from, nil); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err := client.Rcpt(to, nil); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %w", err)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: %s\r\n\r\n%s",
		from, to, subject, contentType(html), body)
	if _, err := io.WriteString(w, msg); err != nil {
		return fmt.Errorf("failed to write email content: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data connection: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("failed to close SMTP connection: %w", err)
	}

	return nil
}

func (s *Sender) SendWithVerifiedSender(from, to, subject, body, replyTo string) error {
	auth := sasl.NewPlainClient("", s.config.SMTPUsername, s.config.SMTPPassword)

	client, err := smtp.DialTLS(fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	if err := client.Mail(from, nil); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err := client.Rcpt(to, nil); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %w", err)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nReply-To: %s\r\n\r\n%s", from, to, subject, replyTo, body)
	if _, err := io.WriteString(w, msg); err != nil {
		return fmt.Errorf("failed to write email content: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data connection: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("failed to close SMTP connection: %w", err)
	}

	return nil
}

type SendRequest struct {
	From    string `json:"from" binding:"required,email"`
	To      string `json:"to" binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body" binding:"required"`
	HTML    bool   `json:"html"`
	ReplyTo string `json:"reply_to"`
}

func (r *SendRequest) Validate() error {
	if len(r.Subject) > 78 {
		return errors.New("subject is too long")
	}
	if len(r.Body) > 1000000 { // 1MB limit
		return errors.New("body is too large")
	}
	return nil
}

func contentType(html bool) string {
	if html {
		return "text/html; charset=UTF-8"
	}
	return "text/plain; charset=UTF-8"
}

func Parse(r io.Reader) (*Email, error) {
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return nil, err
	}

	email := &Email{}

	// Parse From
	if from, err := msg.Header.AddressList("From"); err == nil && len(from) > 0 {
		email.From = from[0].Address
	}

	// Parse To
	if to, err := msg.Header.AddressList("To"); err == nil {
		for _, addr := range to {
			email.To = append(email.To, addr.Address)
		}
	}

	// Parse Subject
	email.Subject = msg.Header.Get("Subject")
	email.Subject, _ = decodeRFC2047(email.Subject)

	// Parse Body
	var bodyBuf bytes.Buffer
	_, err = io.Copy(&bodyBuf, msg.Body)
	if err != nil {
		return nil, err
	}
	email.Body = bodyBuf.String()

	return email, nil
}

func decodeRFC2047(s string) (string, error) {
	dec := new(mime.WordDecoder)
	header, err := dec.DecodeHeader(s)
	if err != nil {
		return s, err
	}
	return header, nil
}