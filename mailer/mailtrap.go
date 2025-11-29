package mailer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/wneessen/go-mail"
)

type mailtrapClient struct {
	fromEmail string
	apiKey    string
}

func NewMailtrapClient(apiKey, fromEmail string) (mailtrapClient, error) {
	if apiKey == "" {
		return mailtrapClient{}, errors.New("api key is required")
	}

	return mailtrapClient{
		fromEmail: fromEmail,
		apiKey:    apiKey,
	}, nil
}

func (m mailtrapClient) SendAPI(templateFile, username, email string, data any, isSandbox bool) error {
	if isSandbox {
		return nil
	}

	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"from": map[string]string{
			"email": m.fromEmail,
			"name":  "NewsDrop",
		},
		"to": []map[string]string{
			{
				"email": email,
				"name":  username,
			},
		},
		"subject": subject.String(),
		"html":    body.String(),
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	bodyReader := bytes.NewReader(jsonPayload)

	url := "https://send.api.mailtrap.io/api/send"
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+m.apiKey)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Mailtrap: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("mailtrap API failed with status %s (%d): %s",
			res.Status, res.StatusCode, string(responseBody))
	}

	return nil
}

func (m mailtrapClient) Send(templateFile, username, email string, data any, isSandbox bool) error {
	if isSandbox {
		return nil
	}

	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return err
	}

	message := mail.NewMsg()
	message.SetAddrHeader("From", m.fromEmail)
	message.SetAddrHeader("To", email)
	message.Subject(subject.String())
	message.AddAlternativeString("text/html", body.String())

	client, err := mail.NewClient("smtp.mailtrap.io", mail.WithPort(587), mail.WithUsername("api"), mail.WithPassword(m.apiKey))
	if err != nil {
		return err
	}

	var retryErr error
	for i := range maxRetries {
		retryErr = client.DialAndSend(message)
		if retryErr != nil {
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to send email after %d attempt, error: %v", maxRetries, retryErr)
}
