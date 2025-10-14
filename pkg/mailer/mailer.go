package mailer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

type Mailer interface {
	SendHTML(to, subject, htmlTpl string, data any) error
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

type mailer struct {
	cfg *Config
}

func New(cfg *Config) Mailer {
	return &mailer{cfg: cfg}
}

func (m *mailer) SendHTML(to, subject, htmlTpl string, data any) error {
	t, err := template.New("email").Parse(htmlTpl)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("exec template: %w", err)
	}

	msg := buildMessage(m.cfg.From, to, subject, body.String())
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	if m.cfg.UseTLS {
		return sendTLS(addr, m.cfg.Host, auth, to, msg)
	}

	return smtp.SendMail(addr, auth, parseFromEmail(m.cfg.From), []string{to}, []byte(msg))
}

func buildMessage(from, to, subject, htmlBody string) string {
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}

	return strings.Join(headers, "\r\n") + "\r\n\r\n" + htmlBody
}

func parseFromEmail(from string) string {
	if i := strings.Index(from, "<"); i >= 0 {
		if j := strings.Index(from[i:], ">"); j > i {
			return strings.TrimSpace(from[i+1 : i+j])
		}
	}

	return strings.TrimSpace(from)
}

func sendTLS(addr, host string, auth smtp.Auth, to string, msg string) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	defer func() {
		if cErr := c.Close(); cErr != nil {
			_ = cErr
		}
	}()

	if err := c.Auth(auth); err != nil {
		return err
	}

	if err := c.Mail(parseFromEmail(addr)); err != nil {
		return err
	}

	if err := c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}

	return w.Close()
}
