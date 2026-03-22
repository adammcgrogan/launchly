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

func (c *Client) SendSitePublished(to, businessName, siteURL string) error {
	html := fmt.Sprintf(`
		<h2>Your site is live — %s!</h2>
		<p>Great news! Your website has been published and is now live. You can view it at the link below:</p>
		<p style="margin:1.5rem 0;">
			<a href="%s" style="display:inline-block;background:#16a34a;color:white;padding:12px 28px;border-radius:8px;text-decoration:none;font-weight:600;font-size:16px;">View Your Site</a>
		</p>
		<p style="color:#666;font-size:13px;">Or copy this link: %s</p>
		<hr>
		<p style="color:#999;font-size:12px;">Sent by AMG Digital</p>
	`, businessName, siteURL, siteURL)
	return c.Send(to, fmt.Sprintf("Your %s website is now live!", businessName), html)
}

func (c *Client) SendSiteUnpublished(to, businessName string) error {
	html := fmt.Sprintf(`
		<h2>Your site has been taken offline — %s</h2>
		<p>Your website has been temporarily unpublished. It is no longer visible to the public.</p>
		<p>If you have any questions or this was unexpected, please get in touch with us.</p>
		<hr>
		<p style="color:#999;font-size:12px;">Sent by AMG Digital</p>
	`, businessName)
	return c.Send(to, fmt.Sprintf("Your %s website has been unpublished", businessName), html)
}

func (c *Client) SendSiteUpdated(to, businessName string, changes []string) error {
	list := ""
	for _, change := range changes {
		list += fmt.Sprintf("<li>%s</li>", change)
	}
	html := fmt.Sprintf(`
		<h2>Your site has been updated — %s</h2>
		<p>The following sections of your website were updated:</p>
		<ul style="margin:1rem 0;padding-left:1.5rem;line-height:2;">%s</ul>
		<p>If you have any questions about the changes, please get in touch.</p>
		<hr>
		<p style="color:#999;font-size:12px;">Sent by AMG Digital</p>
	`, businessName, list)
	return c.Send(to, fmt.Sprintf("Your %s website has been updated", businessName), html)
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
