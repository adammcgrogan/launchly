package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

// wrap puts content inside the standard Launchly email shell.
func wrap(content string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;">

  <!-- Wrapper -->
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f4f5;padding:40px 16px;">
    <tr><td align="center">
      <table width="100%" cellpadding="0" cellspacing="0" style="max-width:560px;">

        <!-- Header bar -->
        <tr>
          <td style="background:#4f46e5;border-radius:12px 12px 0 0;padding:24px 32px;">
            <span style="color:#ffffff;font-size:20px;font-weight:900;letter-spacing:-0.5px;">Launchly</span>
          </td>
        </tr>

        <!-- Body -->
        <tr>
          <td style="background:#ffffff;padding:36px 32px;border-left:1px solid #e5e7eb;border-right:1px solid #e5e7eb;">
            ` + content + `
          </td>
        </tr>

        <!-- Footer -->
        <tr>
          <td style="background:#f9fafb;border:1px solid #e5e7eb;border-top:none;border-radius:0 0 12px 12px;padding:20px 32px;text-align:center;">
            <p style="margin:0;color:#9ca3af;font-size:12px;">
              Sent by <a href="https://launchly.ltd" style="color:#6366f1;text-decoration:none;font-weight:600;">Launchly</a>
              &nbsp;·&nbsp; <a href="mailto:hello@launchly.ltd" style="color:#9ca3af;text-decoration:none;">hello@launchly.ltd</a>
            </p>
          </td>
        </tr>

      </table>
    </td></tr>
  </table>

</body>
</html>`
}

// button renders a full-width CTA button.
func button(href, label, bg string) string {
	return fmt.Sprintf(`
<table width="100%%" cellpadding="0" cellspacing="0" style="margin:28px 0;">
  <tr>
    <td align="center">
      <a href="%s" style="display:inline-block;background:%s;color:#ffffff;padding:14px 32px;border-radius:8px;text-decoration:none;font-weight:700;font-size:15px;">%s</a>
    </td>
  </tr>
</table>`, href, bg, label)
}

// h1 renders the email headline.
func h1(text string) string {
	return fmt.Sprintf(`<h1 style="margin:0 0 16px;font-size:22px;font-weight:800;color:#111827;line-height:1.3;">%s</h1>`, text)
}

// p renders a body paragraph.
func p(text string) string {
	return fmt.Sprintf(`<p style="margin:0 0 16px;font-size:15px;color:#374151;line-height:1.6;">%s</p>`, text)
}

// divider renders a thin horizontal rule.
func divider() string {
	return `<hr style="border:none;border-top:1px solid #e5e7eb;margin:24px 0;">`
}

func (c *Client) SendSitePublished(to, businessName, siteURL string) error {
	content := h1("Your site is live!") +
		p(fmt.Sprintf("Great news — <strong>%s</strong> is now published and visible to the public.", businessName)) +
		p("Click the button below to see your site:") +
		button(siteURL, "View Your Site", "#4f46e5") +
		divider() +
		p(fmt.Sprintf(`<span style="color:#6b7280;font-size:13px;">Or copy this link: <a href="%s" style="color:#4f46e5;">%s</a></span>`, siteURL, siteURL))
	return c.Send(to, fmt.Sprintf("Your %s website is now live!", businessName), wrap(content))
}

func (c *Client) SendSiteUnpublished(to, businessName string) error {
	content := h1("Your site has been taken offline") +
		p(fmt.Sprintf("Your <strong>%s</strong> website has been temporarily unpublished and is no longer visible to the public.", businessName)) +
		p("If you have any questions or this was unexpected, just reply to this email and we'll sort it out.") +
		divider() +
		p(`<span style="color:#6b7280;font-size:13px;">You can get in touch with us any time at <a href="mailto:hello@launchly.ltd" style="color:#4f46e5;">hello@launchly.ltd</a>.</span>`)
	return c.Send(to, fmt.Sprintf("Your %s website has been unpublished", businessName), wrap(content))
}

func (c *Client) SendSiteUpdated(to, businessName string, changes []string) error {
	items := ""
	for _, change := range changes {
		items += fmt.Sprintf(`<li style="margin-bottom:6px;">%s</li>`, change)
	}
	list := fmt.Sprintf(`<ul style="margin:0 0 16px;padding-left:20px;color:#374151;font-size:15px;line-height:1.6;">%s</ul>`, items)

	content := h1("Your site has been updated") +
		p(fmt.Sprintf("We've made the following changes to your <strong>%s</strong> website:", businessName)) +
		list +
		p("The updates are live immediately.") +
		divider() +
		p(`<span style="color:#6b7280;font-size:13px;">Not what you expected? Reply to this email and we'll fix it right away.</span>`)
	return c.Send(to, fmt.Sprintf("Your %s website has been updated", businessName), wrap(content))
}

func (c *Client) SendPaymentLink(to, businessName, checkoutURL string) error {
	content := h1("Your website is ready - complete your payment") +
		p(fmt.Sprintf("Great news! Your <strong>%s</strong> website is built and looking great.", businessName)) +
		p("To go live, complete your payment securely via Stripe. <strong>No charges apply until you're happy.</strong>") +
		button(checkoutURL, "Complete Payment", "#4f46e5") +
		divider() +
		p(fmt.Sprintf(`<span style="color:#6b7280;font-size:13px;">If the button doesn't work, copy and paste this link into your browser:<br><a href="%s" style="color:#4f46e5;word-break:break-all;">%s</a></span>`, checkoutURL, checkoutURL))
	return c.Send(to, fmt.Sprintf("Complete payment for your %s website", businessName), wrap(content))
}

func (c *Client) SendLeadNotification(to, businessName, visitorName, visitorEmail, phone, message string) error {
	rows := ""
	fields := [][2]string{
		{"Name", visitorName},
		{"Email", visitorEmail},
		{"Phone", phone},
	}
	for _, f := range fields {
		if strings.TrimSpace(f[1]) == "" {
			continue
		}
		rows += fmt.Sprintf(`
<tr>
  <td style="padding:10px 14px;font-size:13px;font-weight:600;color:#6b7280;white-space:nowrap;width:80px;">%s</td>
  <td style="padding:10px 14px;font-size:14px;color:#111827;">%s</td>
</tr>`, f[0], f[1])
	}
	if strings.TrimSpace(message) != "" {
		rows += fmt.Sprintf(`
<tr>
  <td style="padding:10px 14px;font-size:13px;font-weight:600;color:#6b7280;vertical-align:top;">Message</td>
  <td style="padding:10px 14px;font-size:14px;color:#111827;">%s</td>
</tr>`, message)
	}

	table := fmt.Sprintf(`
<table width="100%%" cellpadding="0" cellspacing="0" style="border:1px solid #e5e7eb;border-radius:8px;border-collapse:separate;border-spacing:0;overflow:hidden;margin:0 0 24px;">
  %s
</table>`, rows)

	content := h1(fmt.Sprintf("New enquiry - %s", businessName)) +
		p(fmt.Sprintf("Someone just submitted an enquiry through your <strong>%s</strong> website:", businessName)) +
		table +
		divider() +
		p(`<span style="color:#6b7280;font-size:13px;">This lead was submitted through your Launchly website contact form.</span>`)

	return c.Send(to, fmt.Sprintf("New enquiry from your website - %s", businessName), wrap(content))
}
