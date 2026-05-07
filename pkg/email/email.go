package email

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

type smtpConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

var cfg *smtpConfig

func InitSmtpService() {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")
	port := 1025
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		} else {
			log.Printf("端口号 %q 无效,使用默认端口号 1025", portStr)
		}

	}

	cfg = &smtpConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func GenerateVerificationCode() (string, error) {
	newInt := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, newInt)
	if err != nil {
		return "", err
	}
	code := int(n.Int64()) + 100000
	return fmt.Sprintf("%06d", code), nil
}

func SendVerificationCode(to, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "验证码")
	m.SetBody("text/plain", fmt.Sprintf("验证码：%s，5分钟内有效。", code))

	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	return dialer.DialAndSend(m)
}
