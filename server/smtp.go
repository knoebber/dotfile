package server

import (
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"

	"github.com/pkg/errors"
)

// SMTPConfig configures a SMTP client.
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Sender   string `json:"sender"`
}

func smtpConfig(path string) (*SMTPConfig, error) {
	config := new(SMTPConfig)

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "reading smtp config")
	}

	if err = json.Unmarshal(bytes, config); err != nil {
		return nil, errors.Wrapf(err, "unmarshaling smtp config to struct")
	}

	return config, nil
}

func mail(config *SMTPConfig, email, subject, body string) error {
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	to := []string{email}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", email, subject, body))
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	err := smtp.SendMail(addr, auth, config.Sender, to, msg)
	if err != nil {
		return errors.Wrapf(err, "sending mail to %q", email)
	}

	return nil
}
