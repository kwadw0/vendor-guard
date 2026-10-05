package mail

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

var templates = template.Must(template.ParseFS(templateFS,
	"templates/verify_email.html",
	"templates/partner_invite.html",
	"templates/password_reset.html",
))

type verifyData struct {
	Name      string
	Link      string
	ExpiresIn string
}

type inviteData struct {
	PartnerName string
	Link        string
	ExpiresIn   string
}

type resetData struct {
	Name      string
	Link      string
	ExpiresIn string
}

func renderVerifyEmail(name, link string) (subject, html, text string) {
	var buf bytes.Buffer
	_ = templates.ExecuteTemplate(&buf, "verify_email.html", verifyData{
		Name:      name,
		Link:      link,
		ExpiresIn: "24 hours",
	})
	return "Verify your Preuvio email", buf.String(),
		"Hi " + name + ",\n\nVerify your email: " + link + "\n\nThis link expires in 24 hours.\n"
}

func renderPartnerInvite(partnerName, link string) (subject, html, text string) {
	var buf bytes.Buffer
	_ = templates.ExecuteTemplate(&buf, "partner_invite.html", inviteData{
		PartnerName: partnerName,
		Link:        link,
		ExpiresIn:   "7 days",
	})
	return "You've been invited to join " + partnerName + " on Preuvio", buf.String(),
		"You've been invited to join " + partnerName + " on Preuvio.\n\nAccept here: " + link + "\n\nThis link expires in 7 days.\n"
}

func renderPasswordReset(name, link string) (subject, html, text string) {
	var buf bytes.Buffer
	_ = templates.ExecuteTemplate(&buf, "password_reset.html", resetData{
		Name:      name,
		Link:      link,
		ExpiresIn: "1 hour",
	})
	return "Reset your Preuvio password", buf.String(),
		"Hi " + name + ",\n\nReset your password: " + link + "\n\nThis link expires in 1 hour. If you didn't request this, ignore this email.\n"
}
