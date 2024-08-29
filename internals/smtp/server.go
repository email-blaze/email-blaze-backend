package smtp

import (
	"email-blaze/internals/config"
	"email-blaze/internals/email"
	"fmt"
	"io"
	"time"

	gosmtp "github.com/emersion/go-smtp"
)

type Server struct {
	config *config.Config
	sender *email.Sender
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
		sender: email.NewSender(cfg),
	}
}

func (s *Server) Start() error {
	be := &smtpBackend{
		sender: s.sender,
	}
	srv := gosmtp.NewServer(be)

	srv.Addr = fmt.Sprintf(":%d", s.config.SMTPPort)
	srv.Domain = "localhost"
	srv.ReadTimeout = 10 * time.Second
	srv.WriteTimeout = 10 * time.Second
	srv.MaxMessageBytes = 1024 * 1024
	srv.MaxRecipients = 50
	srv.AllowInsecureAuth = true

	return srv.ListenAndServe()
}

type smtpBackend struct {
	sender *email.Sender
}

func (s *smtpBackend) NewSession(state gosmtp.ConnectionState) (gosmtp.Session, error) {
	return &smtpSession{
		sender: s.sender,
	}, nil
}



type smtpSession struct {
	sender *email.Sender
	from   string
	to     []string
}

func (s *smtpSession) AuthPlain(username, password string) error {
	// Implement authentication here
	return nil
}

func (s *smtpSession) Mail(from string, opts *gosmtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *smtpSession) Rcpt(to string) error {
	s.to = append(s.to, to)
	return nil
}

func (s *smtpSession) Data(r io.Reader) error {
	// Implement email handling here
	// Use s.sender to send emails
	return nil
}

func (s *smtpSession) Reset() {
	s.from = ""
	s.to = nil
}

func (s *smtpSession) Logout() error {
	return nil
}