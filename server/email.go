package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/spf13/viper"
	"mime/quotedprintable"
	"net/smtp"
	"time"
)

var emailClient *EmailClient

func GetEmailClient() (*EmailClient, error) {
	if emailClient == nil {
		username := viper.GetString("email.username")
		if username == "" {
			return nil, fmt.Errorf("email username config is empty")
		}
		server := viper.GetString("email.server")
		if server == "" {
			return nil, fmt.Errorf("email server config is empty")
		}
		port := viper.GetInt("email.port")
		if port == 0 {
			return nil, fmt.Errorf("email port config is empty")
		}
		email := viper.GetString("email.email")
		if email == "" {
			return nil, fmt.Errorf("email email config is empty")
		}
		password := viper.GetString("email.password")
		if password == "" {
			return nil, fmt.Errorf("email password config is empty")
		}

		emailClient = &EmailClient{
			username: username,
			server:   server,
			port:     port,
			email:    email,
			password: password,
		}
	}
	newClient := *emailClient
	return &newClient, nil
}

type EmailClient struct {
	username string
	server   string
	port     int
	email    string
	password string

	message           bytes.Buffer
	multipartBoundary string
}

func (e *EmailClient) NewMessage(subject string) *EmailClient {
	e.message = bytes.Buffer{}
	if e.multipartBoundary == "" {
		e.multipartBoundary = fmt.Sprintf("%v", time.Now().UnixNano())
	}

	// Create a MIME multi-part message
	e.message.WriteString(fmt.Sprintf("From: %s<%s>\r\n", e.username, e.email))
	e.message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	e.message.WriteString("MIME-Version: 1.0\r\n")
	e.message.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; charset=UTF-8; boundary=%s\r\n", e.multipartBoundary))
	e.message.WriteString("\r\n")
	return e
}

func (e *EmailClient) AddTextContent(content string) *EmailClient {
	e.message.WriteString(fmt.Sprintf("--%s\r\n", e.multipartBoundary))
	e.message.WriteString("Content-Type: text/plain; charset=\"utf-8\"; format=flowed; delsp=yes\r\n")
	e.message.WriteString("Content-Transfer-Encoding: base64\r\n")
	e.message.WriteString("\r\n")
	encodedTextBody := base64.StdEncoding.EncodeToString([]byte(content))
	e.message.WriteString(encodedTextBody)
	e.message.WriteString("\r\n")
	return e
}

func (e *EmailClient) AddHtmlContent(content string) *EmailClient {
	e.message.WriteString(fmt.Sprintf("--%s\r\n", e.multipartBoundary))
	e.message.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	e.message.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	e.message.WriteString("\r\n")

	encodedHtmlBody, _ := quotedPrintableEncode(content)
	e.message.WriteString(encodedHtmlBody)
	e.message.WriteString("\r\n")
	return e
}

func (e *EmailClient) Send(to []string) error {
	e.message.WriteString(fmt.Sprintf("--%s--\r\n", e.multipartBoundary))

	host := fmt.Sprintf("%s:%d", e.server, e.port)
	auth := smtp.PlainAuth("", e.email, e.password, e.server)
	err := smtp.SendMail(host, auth, e.email, to, e.message.Bytes())
	return err
}

func quotedPrintableEncode(s string) (string, error) {
	var ac bytes.Buffer
	w := quotedprintable.NewWriter(&ac)
	_, err := w.Write([]byte(s))
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}
	return ac.String(), nil
}
