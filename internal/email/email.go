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

func (c *Client) SendPaymentLink(to, businessName, checkoutURL string) error {
	html := fmt.Sprintf(`
		<h2>Your website is ready — %s</h2>
		<p>Great news! We've built your site and it's looking great.</p>
		<p>To complete your order, please click the button below to pay securely via Stripe. You only pay once you're happy with your site.</p>
		<p style="margin:1.5rem 0;">
			<a href="%s" style="display:inline-block;background:#4f46e5;color:white;padding:12px 28px;border-radius:8px;text-decoration:none;font-weight:600;font-size:16px;">Complete Payment</a>
		</p>
		<p style="color:#666;font-size:13px;">If the button doesn't work, copy and paste this link into your browser:<br>%s</p>
		<hr>
		<p style="color:#999;font-size:12px;">Sent by AMG Digital</p>
	`, businessName, checkoutURL, checkoutURL)
	return c.Send(to, fmt.Sprintf("Complete payment for your %s website", businessName), html)
}

func (c *Client) SendLeadNotification(to, businessName, visitorName, visitorEmail, phone, message string) error {
	html := fmt.Sprintf(`
		<h2>New lead for %s</h2>
		<p><strong>Name:</strong> %s</p>
		<p><strong>Email:</strong> %s</p>
		<p><strong>Phone:</strong> %s</p>
		<p><strong>Message:</strong> %s</p>
		<hr>
		<p style="color:#999;font-size:12px;">Sent by AMG Digital</p>
	`, businessName, visitorName, visitorEmail, phone, message)

	return c.Send(to, fmt.Sprintf("New lead from your website — %s", businessName), html)
}
