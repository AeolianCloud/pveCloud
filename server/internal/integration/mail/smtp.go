package mail

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
)

/**
 * Sender 发送用户端账号自助邮件。
 */
type Sender struct {
	cfg config.MailConfig
}

/**
 * NewSender 创建 SMTP 邮件发送器。
 */
func NewSender(cfg config.MailConfig) *Sender {
	return &Sender{cfg: cfg}
}

/**
 * Enabled 返回邮件发送是否启用。
 */
func (s *Sender) Enabled() bool {
	return s != nil && s.cfg.Enabled
}

/**
 * SendPasswordReset 发送密码重置链接。
 */
func (s *Sender) SendPasswordReset(to string, resetURL string) error {
	if !s.Enabled() {
		return fmt.Errorf("mail is disabled")
	}

	subject := "pveCloud 密码重置"
	body := fmt.Sprintf("你正在重置 pveCloud 账号密码。\n\n请打开以下链接完成密码重置：\n%s\n\n如果不是你本人操作，请忽略这封邮件。", resetURL)
	return s.send(to, subject, body)
}

func (s *Sender) send(to string, subject string, body string) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	from := mail.Address{Name: s.cfg.FromName, Address: s.cfg.FromAddress}
	recipients := []string{to}
	message := strings.Join([]string{
		fmt.Sprintf("From: %s", from.String()),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", mime.QEncoding.Encode("UTF-8", subject)),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}
	if s.cfg.UseTLS {
		if s.cfg.Port == 465 {
			return s.sendWithImplicitTLS(addr, auth, recipients, []byte(message))
		}
		return s.sendWithStartTLS(addr, auth, recipients, []byte(message))
	}
	return smtp.SendMail(addr, auth, s.cfg.FromAddress, recipients, []byte(message))
}

func (s *Sender) sendWithImplicitTLS(addr string, auth smtp.Auth, recipients []string, message []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return err
	}
	defer client.Quit()

	return s.sendWithClient(client, auth, recipients, message)
}

func (s *Sender) sendWithStartTLS(addr string, auth smtp.Auth, recipients []string, message []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Quit()

	if err := client.StartTLS(&tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
		return err
	}
	return s.sendWithClient(client, auth, recipients, message)
}

func (s *Sender) sendWithClient(client *smtp.Client, auth smtp.Auth, recipients []string, message []byte) error {
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(s.cfg.FromAddress); err != nil {
		return err
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}
