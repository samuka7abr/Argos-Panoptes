package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type Notifier struct {
	config EmailConfig
}

func NewNotifier(config EmailConfig) *Notifier {
	return &Notifier{config: config}
}

func (n *Notifier) SendAlert(rule *Rule, value float64) error {
	subject := fmt.Sprintf("[%s] %s", strings.ToUpper(rule.Severity), rule.Name)
	body := n.buildEmailBody(rule, value)

	for _, recipient := range rule.EmailTo {
		if err := n.sendEmail(recipient, subject, body); err != nil {
			return fmt.Errorf("failed to send to %s: %w", recipient, err)
		}
	}

	return nil
}

func (n *Notifier) buildEmailBody(rule *Rule, value float64) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Alert: %s\n", rule.Name))
	sb.WriteString(fmt.Sprintf("Severity: %s\n", rule.Severity))
	sb.WriteString(fmt.Sprintf("Time: %s\n\n", time.Now().Format(time.RFC3339)))

	sb.WriteString(fmt.Sprintf("Description: %s\n\n", rule.Description))

	sb.WriteString(fmt.Sprintf("Expression: %s\n", rule.Expr))
	sb.WriteString(fmt.Sprintf("Current Value: %.2f\n\n", value))

	if rule.Service != "" {
		sb.WriteString(fmt.Sprintf("Service: %s\n", rule.Service))
	}
	if rule.Target != "" {
		sb.WriteString(fmt.Sprintf("Target: %s\n", rule.Target))
	}

	sb.WriteString("\n---\n")
	sb.WriteString("Argos Panoptes Monitoring System\n")

	return sb.String()
}

func (n *Notifier) sendEmail(to, subject, body string) error {
	from := n.config.From
	password := n.config.SMTPPassword
	smtpHost := n.config.SMTPHost
	smtpPort := n.config.SMTPPort

	message := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, to, subject, body))

	auth := smtp.PlainAuth("", n.config.SMTPUser, password, smtpHost)

	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)

	if n.config.UseTLS {
		return n.sendEmailTLS(addr, auth, from, []string{to}, message)
	}

	return smtp.SendMail(addr, auth, from, []string{to}, message)
}

func (n *Notifier) sendEmailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.StartTLS(&tls.Config{ServerName: n.config.SMTPHost}); err != nil {
		return err
	}

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}
