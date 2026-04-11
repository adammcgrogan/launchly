package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/adammcgrogan/launchly/internal/models"
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
	ReplyTo []string `json:"reply_to,omitempty"`
}

func (c *Client) Send(to, subject, html string) error {
	return c.sendWithReplyTo(to, subject, html, "")
}

func (c *Client) sendWithReplyTo(to, subject, html, replyTo string) error {
	payload := sendRequest{
		From:    c.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}
	if replyTo != "" {
		payload.ReplyTo = []string{replyTo}
	}
	body, err := json.Marshal(payload)
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

func (c *Client) SendWelcomeEmail(to, businessName string) error {
	content := h1("We've received your details!") +
		p(fmt.Sprintf("Thanks for submitting your information for <strong>%s</strong>. We're on it!", businessName)) +
		p("Here's what happens next:") +
		`<ol style="margin:0 0 16px;padding-left:20px;color:#374151;font-size:15px;line-height:1.8;">
  <li>We review your details and build your site</li>
  <li>You'll receive an email with your site link within 24 hours</li>
  <li>Once you're happy, your site goes live and starts attracting customers</li>
</ol>` +
		p("If you have any questions in the meantime, just reply to this email.") +
		divider() +
		p(`<span style="color:#6b7280;font-size:13px;">Questions? Contact us at <a href="mailto:hello@launchly.ltd" style="color:#4f46e5;">hello@launchly.ltd</a></span>`)
	return c.Send(to, fmt.Sprintf("We've received your details — %s", businessName), wrap(content))
}

func (c *Client) SendPaymentConfirmation(to, businessName, plan string) error {
	planLabel := "Starter"
	if plan == "pro" {
		planLabel = "Pro"
	}
	content := h1("Payment confirmed — you're all set!") +
		p(fmt.Sprintf("Thanks for subscribing to the <strong>%s plan</strong> for <strong>%s</strong>.", planLabel, businessName)) +
		p("Your site is now live and active. Enquiries from your site will be forwarded straight to your inbox.") +
		divider() +
		p(`<span style="color:#6b7280;font-size:13px;">Need to make changes? Just reply to this email and we'll take care of it.</span>`)
	return c.Send(to, fmt.Sprintf("Payment confirmed for %s", businessName), wrap(content))
}

func (c *Client) SendCancellationConfirmation(to, businessName string) error {
	content := h1("Your subscription has been cancelled") +
		p(fmt.Sprintf("We've cancelled the subscription for <strong>%s</strong>. Your site will be taken offline shortly.", businessName)) +
		p("If this was a mistake or you'd like to reactivate your site in future, just get in touch.") +
		divider() +
		p(`<span style="color:#6b7280;font-size:13px;">We're sorry to see you go. Contact us at <a href="mailto:hello@launchly.ltd" style="color:#4f46e5;">hello@launchly.ltd</a></span>`)
	return c.Send(to, fmt.Sprintf("Subscription cancelled — %s", businessName), wrap(content))
}

func (c *Client) SendAnalyticsDigest(to, businessName, frequency string, stats *models.SiteStats, siteURL string) error {
	period := "weekly"
	days := "7 days"
	if frequency == "monthly" {
		period = "monthly"
		days = "30 days"
	}

	statsRow := fmt.Sprintf(`
<table width="100%%" cellpadding="0" cellspacing="0" style="margin:0 0 24px;">
  <tr>
    <td width="50%%" style="padding:0 8px 0 0;">
      <div style="background:#f9fafb;border:1px solid #e5e7eb;border-radius:8px;padding:20px;text-align:center;">
        <div style="font-size:36px;font-weight:900;color:#111827;line-height:1;">%d</div>
        <div style="font-size:13px;color:#6b7280;margin-top:4px;">Total visits</div>
      </div>
    </td>
    <td width="50%%" style="padding:0 0 0 8px;">
      <div style="background:#f9fafb;border:1px solid #e5e7eb;border-radius:8px;padding:20px;text-align:center;">
        <div style="font-size:36px;font-weight:900;color:#111827;line-height:1;">%d</div>
        <div style="font-size:13px;color:#6b7280;margin-top:4px;">Unique visitors</div>
      </div>
    </td>
  </tr>
</table>`, stats.TotalViews, stats.UniqueVisitors)

	var daysTable string
	if len(stats.ViewsByDay) > 0 {
		rows := ""
		for _, d := range stats.ViewsByDay {
			rows += fmt.Sprintf(`<tr>
  <td style="padding:7px 14px;font-size:13px;color:#374151;border-bottom:1px solid #f3f4f6;">%s</td>
  <td style="padding:7px 14px;font-size:13px;font-weight:700;color:#111827;border-bottom:1px solid #f3f4f6;text-align:right;">%d</td>
</tr>`, d.Day.Format("Mon 2 Jan"), d.Count)
		}
		daysTable = fmt.Sprintf(`<p style="margin:0 0 8px;font-size:12px;font-weight:700;color:#9ca3af;text-transform:uppercase;letter-spacing:.07em;">Views by day</p>
<table width="100%%" cellpadding="0" cellspacing="0" style="border:1px solid #e5e7eb;border-radius:8px;overflow:hidden;margin:0 0 24px;border-collapse:separate;border-spacing:0;">%s</table>`, rows)
	}

	var refTable string
	if len(stats.TopReferrers) > 0 {
		rows := ""
		for _, ref := range stats.TopReferrers {
			label := ref.Referrer
			if label == "" {
				label = "Direct / unknown"
			}
			rows += fmt.Sprintf(`<tr>
  <td style="padding:7px 14px;font-size:13px;color:#374151;border-bottom:1px solid #f3f4f6;">%s</td>
  <td style="padding:7px 14px;font-size:13px;font-weight:700;color:#111827;border-bottom:1px solid #f3f4f6;text-align:right;">%d</td>
</tr>`, label, ref.Count)
		}
		refTable = fmt.Sprintf(`<p style="margin:0 0 8px;font-size:12px;font-weight:700;color:#9ca3af;text-transform:uppercase;letter-spacing:.07em;">Where visitors came from</p>
<table width="100%%" cellpadding="0" cellspacing="0" style="border:1px solid #e5e7eb;border-radius:8px;overflow:hidden;margin:0 0 24px;border-collapse:separate;border-spacing:0;">%s</table>`, rows)
	}

	noDataNote := ""
	if stats.TotalViews == 0 {
		noDataNote = p(`<span style="color:#6b7280;">No visits were recorded in this period. Once your site gets traffic, you'll see a full breakdown here.</span>`)
	}

	content := h1(fmt.Sprintf("Your %s website report", period)) +
		p(fmt.Sprintf("Here's how <strong>%s</strong> performed over the last %s.", businessName, days)) +
		statsRow +
		noDataNote +
		daysTable +
		refTable +
		button(siteURL, "View Your Website", "#4f46e5") +
		divider() +
		p(`<span style="color:#6b7280;font-size:13px;">You're receiving this report because analytics is enabled for your site. To change your report frequency, contact us at <a href="mailto:hello@launchly.ltd" style="color:#4f46e5;">hello@launchly.ltd</a>.</span>`)

	subject := fmt.Sprintf("Your weekly website report — %s", businessName)
	if frequency == "monthly" {
		subject = fmt.Sprintf("Your monthly website report — %s", businessName)
	}
	return c.Send(to, subject, wrap(content))
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

	return c.sendWithReplyTo(to, fmt.Sprintf("New enquiry from your website - %s", businessName), wrap(content), visitorEmail)
}
