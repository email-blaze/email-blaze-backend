package smtp

import (
	"bytes"
	"email-blaze/internals/auth"
	"email-blaze/internals/config"
	"email-blaze/internals/email"
	"email-blaze/internals/logger"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"crypto/tls"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

type Backend struct {
	config *config.Config
	sender *email.Sender
}

func NewBackend(cfg *config.Config, sender *email.Sender) *Backend {
	return &Backend{
		config: cfg,
		sender: sender,
	}
}

func (bkd *Backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &Session{
		backend: bkd,
	}, nil
}

type Session struct {
	backend *Backend
	from    string
	to      []string
}

func (s *Session) AuthMechanisms() []string {
	return []string{sasl.Plain}
}

func (s *Session) Auth(mech string) (sasl.Server, error) {
	return sasl.NewPlainServer(func(identity, username, password string) error {
		logger.Info("Auth attempt",
            logger.Field("identity", identity),
            logger.Field("username", username),
            logger.Field("password", password),
            logger.Field("config_username", s.backend.config.SMTPUsername),
            logger.Field("config_password", s.backend.config.SMTPPassword))
		if username != s.backend.config.SMTPUsername || password != s.backend.config.SMTPPassword {
			logger.Error("Authentication failed",
			logger.Field("provided_username", username),
			logger.Field("config_username", s.backend.config.SMTPUsername))
			return fmt.Errorf("invalid username or password")
		}
		return nil
	}), nil
}

func (s *Session) Mail(from string, _ *smtp.MailOptions) error {
	isValid, err := auth.VerifyEmail(from)
	if err != nil {
		return fmt.Errorf("email verification failed: %w", err)
	}
	if !isValid {
			return errors.New("invalid email or domain not properly configured")
	}
	s.from = from
	return nil
}

func (s *Session) Rcpt(to string, _ *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	parsedEmail, err := email.Parse(bytes.NewReader(b))
	if err != nil {
		return err
	}

	// Determine if the email is HTML
	isHTML := false
	if strings.Contains(string(b), "Content-Type: text/html") {
		isHTML = true
	}

	for _, recipient := range s.to {
		err := s.backend.sender.Send(s.from, recipient, parsedEmail.Subject, parsedEmail.Body, isHTML, s.backend.config.DefaultUser.Domain)
		if err != nil {
			logger.Error("Failed to send email", logger.Field("error", err))
			return err
		}
	}

	return nil
}

func (s *Session) Reset() {
	s.from = ""
	s.to = nil
}

func (s *Session) Logout() error {
	return nil
}

func StartSMTPServer(cfg *config.Config, sender *email.Sender) error {
	be := NewBackend(cfg, sender)
	s := smtp.NewServer(be)

	s.Addr = fmt.Sprintf(":%d", cfg.SMTPPort)
	s.Domain = cfg.SMTPHost
	s.ReadTimeout = 30 * time.Second
	s.WriteTimeout = 30 * time.Second
	s.MaxMessageBytes = 10 * 1024 * 1024 // 10 MB
	s.MaxRecipients = 50
	s.AllowInsecureAuth = cfg.DevelopmentMode // Disable insecure authentication


	if !cfg.DevelopmentMode {
	// Configure TLS
	cert, err := tls.LoadX509KeyPair(cfg.SSLCertFile, cfg.SSLKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %w", err)
	}
	s.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	s.EnableSMTPUTF8 = true
	s.EnableREQUIRETLS = true
	s.EnableBINARYMIME = true
	s.EnableDSN = true
	// Implement rate limiting
	s.MaxLineLength = 1000

		logger.Info("Starting SMTP server in development mode (TLS disabled)", logger.Field("addr", s.Addr))
		return s.ListenAndServeTLS() // Use ListenAndServeTLS for secure connections
	} else {
		logger.Info("Starting SMTP server", logger.Field("addr", s.Addr))
		return s.ListenAndServe() // Use ListenAndServe for insecure connections
	}
}
