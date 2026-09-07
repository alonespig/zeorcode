package mail

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Mailer 极简 SMTP 发信器：465 走隐式 TLS，其余端口(587/25)走 STARTTLS。
// 不引第三方库，够验证码这种单收件人 HTML 邮件用。
type Mailer struct {
	host     string
	port     int
	username string
	password string
	fromName string
}

func NewMailer() *Mailer {
	return &Mailer{
		host:     viper.GetString("mail.host"),
		port:     viper.GetInt("mail.port"),
		username: viper.GetString("mail.username"),
		password: viper.GetString("mail.password"),
		fromName: viper.GetString("mail.from_name"),
	}
}

// Enabled 是否配置了发信账号；未配置时上层可给出「邮件服务未配置」的友好提示。
func (m *Mailer) Enabled() bool {
	return m.host != "" && m.username != "" && m.password != ""
}

// Send 发送一封 UTF-8 HTML 邮件（单收件人）。
func (m *Mailer) Send(to, subject, htmlBody string) error {
	if !m.Enabled() {
		return fmt.Errorf("邮件服务未配置(mail.host/username/password)")
	}

	from := m.username
	fromHeader := from
	if m.fromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("UTF-8", m.fromName), from)
	}

	var b strings.Builder
	b.WriteString("From: " + fromHeader + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("UTF-8", subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(htmlBody)
	msg := []byte(b.String())

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	if m.port == 465 {
		return m.sendImplicitTLS(addr, auth, from, to, msg)
	}
	// 587/25：明文连接 + STARTTLS，交给标准库处理
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

// sendImplicitTLS 处理 465 端口的隐式 TLS：先建立 TLS 连接，再跑 SMTP 会话。
func (m *Mailer) sendImplicitTLS(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 15 * time.Second},
		"tcp", addr, &tls.Config{ServerName: m.host},
	)
	if err != nil {
		return fmt.Errorf("连接邮件服务器失败: %w", err)
	}
	_ = conn.SetDeadline(time.Now().Add(20 * time.Second))

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer func() { _ = client.Close() }()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败(检查授权码): %w", err)
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
