package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	apiKey string
	from   string
}

func New(apiKey, from string) *Client {
	return &Client{apiKey: apiKey, from: from}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (c *Client) Send(to, subject, html string) error {
	body, err := json.Marshal(sendRequest{
		From:    c.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: %s", resp.Status)
	}
	return nil
}

func (c *Client) SendLeadNotification(to, businessName, visitorName, phone, message string) error {
	html := fmt.Sprintf(`
		<h2>New lead for %s</h2>
		<p><strong>Name:</strong> %s</p>
		<p><strong>Phone:</strong> %s</p>
		<p><strong>Message:</strong> %s</p>
		<hr>
		<p style="color:#999;font-size:12px;">Sent by LocalLaunch</p>
	`, businessName, visitorName, phone, message)

	return c.Send(to, fmt.Sprintf("New lead from your website — %s", businessName), html)
}
