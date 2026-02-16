package common

import (
	"fmt"
	"net/smtp"
	"gsupert/internal/config"
)

type EmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	from := s.cfg.SMTPFrom
	fromName := s.cfg.SMTPFromName

	// Ensure headers follow RFC 2822
	// Use double quotes for name to avoid syntax errors if it contains spaces
	msg := fmt.Sprintf("MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=\"utf-8\"\r\n"+
		"From: \"%s\" <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s", fromName, from, to, subject, body)

	// Auth can be nil for Mailpit local testing
	var auth smtp.Auth
	if s.cfg.SMTPUser != "" && s.cfg.SMTPPass != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	}

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}
