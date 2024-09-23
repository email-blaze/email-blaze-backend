package smtp

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"email-blaze/internals/auth"
	"email-blaze/internals/config"
	"email-blaze/internals/email"
	"email-blaze/internals/logger"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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
		id:      generateUniqueID(),
		backend: bkd,
	}, nil
}

func generateUniqueID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

type Session struct {
	id      string
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
		logger.Error("Email verification failed", logger.Err(err))
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
	logger.Info("Received email data", logger.Field("sessionID", s.id), logger.Field("from", s.from), logger.Field("to", s.to))
	
	var b bytes.Buffer
	reader := bufio.NewReader(r)
	
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("EOF reached", logger.Field("sessionID", s.id))
				break
			}
			logger.Error("Failed to read email data", logger.Field("sessionID", s.id), logger.Err(err))
			return fmt.Errorf("failed to read email data: %w", err)
		}

		// Check for end-of-data marker
		if line == ".\r\n" || line == ".\n" {
			logger.Info("End-of-data marker found", logger.Field("sessionID", s.id))
			break
		}

		// Remove leading dot if line starts with ".."
		if strings.HasPrefix(line, "..") {
			line = line[1:]
		}

		if _, err := b.WriteString(line); err != nil {
			logger.Error("Failed to write to buffer", logger.Field("sessionID", s.id), logger.Err(err))
			return fmt.Errorf("failed to write to buffer: %w", err)
		}

		if b.Len() > s.backend.config.MaxMessageSize {
			logger.Error("Message too large", logger.Field("sessionID", s.id), logger.Field("size", b.Len()))
			return errors.New("message too large")
		}
	}

	// Process the email
	err := s.processEmail(&b)
	if err != nil {
		logger.Error("Failed to process email", logger.Field("sessionID", s.id), logger.Err(err))
		return err
	}

	logger.Info("Email processed successfully", logger.Field("sessionID", s.id))
	return nil
}

func (s *Session) processEmail(b *bytes.Buffer) error {
	start := time.Now()
	parsedEmail, err := email.Parse(bytes.NewReader(b.Bytes()))
	if err != nil {
		logger.Error("Failed to parse email", logger.Field("sessionID", s.id), logger.Err(err))
		return fmt.Errorf("failed to parse email: %w", err)
	}
	parseTime := time.Since(start)

	// Determine if the email is HTML
	isHTML := strings.Contains(b.String(), "Content-Type: text/html")

	for _, recipient := range s.to {
		sendStart := time.Now()
		err := s.backend.sender.Send(s.from, recipient, parsedEmail.Subject, parsedEmail.Body, isHTML, s.backend.config.DefaultUser.Domain)
		sendTime := time.Since(sendStart)
		if err != nil {
			logger.Error("Failed to send email", 
				logger.Field("sessionID", s.id),
				logger.Field("error", err),
				logger.Field("from", s.from),
				logger.Field("to", recipient),
				logger.Field("subject", parsedEmail.Subject),
				logger.Field("parseTime", parseTime),
				logger.Field("sendTime", sendTime))
			return fmt.Errorf("failed to send email: %w", err)
		}
		logger.Info("Email sent successfully", 
			logger.Field("sessionID", s.id),
			logger.Field("from", s.from),
			logger.Field("to", recipient),
			logger.Field("subject", parsedEmail.Subject),
			logger.Field("parseTime", parseTime),
			logger.Field("sendTime", sendTime))
	}

	totalTime := time.Since(start)
	logger.Info("Email processed successfully", 
		logger.Field("sessionID", s.id),
		logger.Field("from", s.from),
		logger.Field("to", s.to),
		logger.Field("subject", parsedEmail.Subject),
		logger.Field("size", b.Len()),
		logger.Field("isHTML", isHTML),
		logger.Field("totalTime", totalTime))

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
	s.ReadTimeout = time.Duration(cfg.SMTPReadTimeout) * time.Second
	s.WriteTimeout = time.Duration(cfg.SMTPWriteTimeout) * time.Second
	s.MaxMessageBytes = int64(cfg.MaxMessageSize)
	s.MaxRecipients = cfg.MaxRecipients
	s.AllowInsecureAuth = cfg.DevelopmentMode

	s.EnableSMTPUTF8 = true
	s.EnableBINARYMIME = true
	s.EnableDSN = true
	s.MaxLineLength = cfg.MaxLineLength

	logger.Info("Starting SMTP server", 
		logger.Field("addr", s.Addr),
		logger.Field("domain", s.Domain),
		logger.Field("allowInsecureAuth", s.AllowInsecureAuth),
		logger.Field("maxMessageBytes", s.MaxMessageBytes),
		logger.Field("maxRecipients", s.MaxRecipients))

	if cfg.DevelopmentMode {
		s.Debug = os.Stdout
		return s.ListenAndServe()
	} else {
		cert, err := tls.LoadX509KeyPair(cfg.SSLCertFile, cfg.SSLKeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificate: %w", err)
		}
		s.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
		s.EnableREQUIRETLS = true
		return s.ListenAndServeTLS()
	}
}
