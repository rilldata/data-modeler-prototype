package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"math"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

//go:embed templates/gen/*
var templatesFS embed.FS

const dateFormat = "January 2, 2006"

type Client struct {
	Sender    Sender
	templates *template.Template
}

func New(sender Sender) *Client {
	return &Client{
		Sender:    sender,
		templates: template.Must(template.New("").ParseFS(templatesFS, "templates/gen/*.html")),
	}
}

type ScheduledReport struct {
	ToEmail        string
	ToName         string
	DisplayName    string
	ReportTime     time.Time
	DownloadFormat string
	OpenLink       string
	DownloadLink   string
	EditLink       string
	External       bool
}

type scheduledReportData struct {
	DisplayName      string
	ReportTimeString string // Will be inferred from ReportTime
	DownloadFormat   string
	OpenLink         template.URL
	DownloadLink     template.URL
	EditLink         template.URL
	External         bool
}

func (c *Client) SendScheduledReport(opts *ScheduledReport) error {
	// Build template data
	data := &scheduledReportData{
		DisplayName:      opts.DisplayName,
		ReportTimeString: opts.ReportTime.Format(time.RFC1123),
		DownloadFormat:   opts.DownloadFormat,
		OpenLink:         template.URL(opts.OpenLink),
		DownloadLink:     template.URL(opts.DownloadLink),
		EditLink:         template.URL(opts.EditLink),
		External:         opts.External,
	}

	// Build subject
	subject := fmt.Sprintf("%s (%s)", opts.DisplayName, data.ReportTimeString)

	var err error
	// Resolve template
	buf := new(bytes.Buffer)
	err = c.templates.Lookup("scheduled_report.html").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()

	return c.Sender.Send(opts.ToEmail, opts.ToName, subject, html)
}

func (c *Client) SendAlertStatus(opts *drivers.AlertStatus) error {
	switch opts.Status {
	case runtimev1.AssertionStatus_ASSERTION_STATUS_PASS:
		return c.sendAlertStatus(opts, &alertStatusData{
			DisplayName:         opts.DisplayName,
			ExecutionTimeString: opts.ExecutionTime.Format(time.RFC1123),
			IsPass:              true,
			IsRecover:           opts.IsRecover,
			OpenLink:            template.URL(opts.OpenLink),
			EditLink:            template.URL(opts.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL:
		return c.sendAlertFail(opts, &alertFailData{
			DisplayName:         opts.DisplayName,
			ExecutionTimeString: opts.ExecutionTime.Format(time.RFC1123),
			FailRow:             opts.FailRow,
			OpenLink:            template.URL(opts.OpenLink),
			EditLink:            template.URL(opts.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR:
		return c.sendAlertStatus(opts, &alertStatusData{
			DisplayName:         opts.DisplayName,
			ExecutionTimeString: opts.ExecutionTime.Format(time.RFC1123),
			IsError:             true,
			ErrorMessage:        opts.ExecutionError,
			OpenLink:            template.URL(opts.EditLink), // NOTE: Using edit link here since for errors, we don't want to open a dashboard, but rather the alert itself
			EditLink:            template.URL(opts.EditLink),
		})
	default:
		return fmt.Errorf("unknown assertion status: %v", opts.Status)
	}
}

type alertFailData struct {
	DisplayName         string
	ExecutionTimeString string // Will be inferred from ExecutionTime
	FailRow             map[string]any
	OpenLink            template.URL
	EditLink            template.URL
}

func (c *Client) sendAlertFail(opts *drivers.AlertStatus, data *alertFailData) error {
	subject := fmt.Sprintf("%s (%s)", data.DisplayName, data.ExecutionTimeString)

	buf := new(bytes.Buffer)
	err := c.templates.Lookup("alert_fail.html").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()

	return c.Sender.Send(opts.ToEmail, opts.ToName, subject, html)
}

type alertStatusData struct {
	DisplayName         string
	ExecutionTimeString string // Will be inferred from ExecutionTime
	IsPass              bool
	IsRecover           bool
	IsError             bool
	ErrorMessage        string
	OpenLink            template.URL
	EditLink            template.URL
}

func (c *Client) sendAlertStatus(opts *drivers.AlertStatus, data *alertStatusData) error {
	subject := fmt.Sprintf("%s (%s)", data.DisplayName, data.ExecutionTimeString)
	if data.IsRecover {
		subject = fmt.Sprintf("Recovered: %s", subject)
	}

	buf := new(bytes.Buffer)
	err := c.templates.Lookup("alert_status.html").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()

	return c.Sender.Send(opts.ToEmail, opts.ToName, subject, html)
}

type CallToAction struct {
	ToEmail    string
	ToName     string
	Subject    string
	PreButton  template.HTML
	ButtonText string
	ButtonLink string
	PostButton template.HTML
	ShowFooter bool
}

func (c *Client) SendCallToAction(opts *CallToAction) error {
	buf := new(bytes.Buffer)
	err := c.templates.Lookup("call_to_action.html").Execute(buf, opts)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()
	return c.Sender.Send(opts.ToEmail, opts.ToName, opts.Subject, html)
}

type Informational struct {
	ToEmail    string
	ToName     string
	Subject    string
	Body       template.HTML
	ShowFooter bool
}

func (c *Client) SendInformational(opts *Informational) error {
	buf := new(bytes.Buffer)
	err := c.templates.Lookup("informational.html").Execute(buf, opts)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()
	return c.Sender.Send(opts.ToEmail, opts.ToName, opts.Subject, html)
}

type Welcome struct {
	ToEmail     string
	ToName      string
	Subject     string
	FrontendURL string
	WelcomeText template.HTML
}

func (c *Client) SendWelcomeToTrial(opts *Welcome) error {
	buf := new(bytes.Buffer)
	err := c.templates.Lookup("welcome_to_trial.html").Execute(buf, opts)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()
	return c.Sender.Send(opts.ToEmail, opts.ToName, opts.Subject, html)
}

func (c *Client) SendWelcomeToTeam(opts *Welcome) error {
	buf := new(bytes.Buffer)
	err := c.templates.Lookup("welcome_to_team.html").Execute(buf, opts)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()
	return c.Sender.Send(opts.ToEmail, opts.ToName, opts.Subject, html)
}

type OrganizationInvite struct {
	ToEmail       string
	ToName        string
	AcceptURL     string
	OrgName       string
	RoleName      string
	InvitedByName string
}

func (c *Client) SendOrganizationInvite(opts *OrganizationInvite) error {
	if opts.InvitedByName == "" {
		opts.InvitedByName = "Rill"
	}

	return c.SendCallToAction(&CallToAction{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("%s invited you to join Rill", opts.InvitedByName),
		PreButton:  template.HTML(fmt.Sprintf("%s has invited you to join <b>%s</b> as a %s for their Rill account. Get started interacting with fast, exploratory dashboards by clicking the button below to sign in and accept your invitation.", opts.InvitedByName, opts.OrgName, opts.RoleName)),
		ButtonText: "Accept invitation",
		ButtonLink: opts.AcceptURL,
	})
}

type OrganizationAddition struct {
	ToEmail       string
	ToName        string
	OpenURL       string
	OrgName       string
	RoleName      string
	InvitedByName string
}

func (c *Client) SendOrganizationAddition(opts *OrganizationAddition) error {
	if opts.InvitedByName == "" {
		opts.InvitedByName = "Rill"
	}

	return c.SendCallToAction(&CallToAction{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("%s has added you to %s", opts.InvitedByName, opts.OrgName),
		PreButton:  template.HTML(fmt.Sprintf("%s has added you as a %s for <b>%s</b>. Click the button below to view and collaborate on Rill dashboard projects for %s.", opts.InvitedByName, opts.RoleName, opts.OrgName, opts.OrgName)),
		ButtonText: "View account",
		ButtonLink: opts.OpenURL,
	})
}

type ProjectInvite struct {
	ToEmail       string
	ToName        string
	AcceptURL     string
	OrgName       string
	ProjectName   string
	RoleName      string
	InvitedByName string
}

func (c *Client) SendProjectInvite(opts *ProjectInvite) error {
	if opts.InvitedByName == "" {
		opts.InvitedByName = "Rill"
	}

	return c.SendCallToAction(&CallToAction{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("You have been invited to the %s/%s project", opts.OrgName, opts.ProjectName),
		PreButton:  template.HTML(fmt.Sprintf("%s has invited you to collaborate as a %s for the <b>%s/%s</b> project. Click the button below to accept your invitation. ", opts.InvitedByName, opts.RoleName, opts.OrgName, opts.ProjectName)),
		ButtonText: "Accept invitation",
		ButtonLink: opts.AcceptURL,
	})
}

type ProjectAddition struct {
	ToEmail       string
	ToName        string
	OpenURL       string
	OrgName       string
	ProjectName   string
	RoleName      string
	InvitedByName string
}

func (c *Client) SendProjectAddition(opts *ProjectAddition) error {
	if opts.InvitedByName == "" {
		opts.InvitedByName = "Rill"
	}

	return c.SendCallToAction(&CallToAction{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("You have been added to the %s/%s project", opts.OrgName, opts.ProjectName),
		PreButton:  template.HTML(fmt.Sprintf("%s has invited you to collaborate as a %s for the <b>%s</b> project. Click the button below to accept your invitation. ", opts.InvitedByName, opts.RoleName, opts.ProjectName)),
		ButtonText: "View account",
		ButtonLink: opts.OpenURL,
	})
}

type ProjectAccessRequest struct {
	Title       string
	Body        template.HTML
	ToEmail     string
	ToName      string
	Email       string
	OrgName     string
	ProjectName string
	ApproveLink string
	DenyLink    string
}

func (c *Client) SendProjectAccessRequest(opts *ProjectAccessRequest) error {
	subject := fmt.Sprintf("%s would like to view %s/%s", opts.Email, opts.OrgName, opts.ProjectName)
	if opts.Body == "" {
		opts.Body = template.HTML(fmt.Sprintf("<b>%s</b> would like to view <b>%s/%s</b>", opts.Email, opts.OrgName, opts.ProjectName))
	}

	buf := new(bytes.Buffer)
	err := c.templates.Lookup("project_access_request.html").Execute(buf, opts)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()
	return c.Sender.Send(opts.ToEmail, opts.ToName, subject, html)
}

type ProjectAccessGranted struct {
	ToEmail     string
	ToName      string
	OpenURL     string
	OrgName     string
	ProjectName string
}

func (c *Client) SendProjectAccessGranted(opts *ProjectAccessGranted) error {
	return c.SendCallToAction(&CallToAction{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("Your request to %s/%s has been approved", opts.OrgName, opts.ProjectName),
		PreButton:  template.HTML(fmt.Sprintf("Your request to <b>%s/%s</b> has been approved", opts.OrgName, opts.ProjectName)),
		ButtonText: "View project in Rill",
		ButtonLink: opts.OpenURL,
	})
}

type ProjectAccessRejected struct {
	ToEmail     string
	ToName      string
	OrgName     string
	ProjectName string
}

func (c *Client) SendProjectAccessRejected(opts *ProjectAccessRejected) error {
	return c.SendInformational(&Informational{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Your request to %s/%s has been denied", opts.OrgName, opts.ProjectName),
		Body:    template.HTML(fmt.Sprintf("Your request to <b>%s/%s</b> has been denied. Contact your project admin for help.", opts.OrgName, opts.ProjectName)),
	})
}

type InvoicePaymentFailed struct {
	ToEmail            string
	ToName             string
	OrgName            string
	Currency           string
	Amount             string
	PaymentURL         string
	GracePeriodEndDate time.Time
}

func (c *Client) SendInvoicePaymentFailed(opts *InvoicePaymentFailed) error {
	return c.SendCallToAction(&CallToAction{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Payment failed for %s. Please update your payment method", opts.OrgName),
		PreButton: template.HTML(fmt.Sprintf(`
We couldn’t process your payment for <b>%s</b>. You have until <b>%s</b> to update your payment details before your org is hibernated.
`, opts.OrgName, opts.GracePeriodEndDate.Format(dateFormat))),
		ButtonText: "Update Payment Info",
		ButtonLink: opts.PaymentURL,
		ShowFooter: true,
	})
}

type InvoicePaymentSuccess struct {
	ToEmail        string
	ToName         string
	OrgName        string
	PaymentDate    time.Time
	BillingPageURL string
}

// SendInvoicePaymentSuccess Currently Used only when a previously failed invoice payment succeeds
func (c *Client) SendInvoicePaymentSuccess(opts *InvoicePaymentSuccess) error {
	return c.SendInformational(&Informational{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Successful payment %s", opts.PaymentDate.Format(dateFormat)),
		Body: template.HTML(fmt.Sprintf(`
Thank you for your payment!
<br /><br />
Your payment for <b>%s</b> has been successfully processed. 
<br /><br />
If you believe this charge to be in error or have any questions, please email support@rilldata.com.
<br /><br />
You can manage your subscription by visiting the <a href=%q>Billing settings</a>
`, opts.OrgName, opts.BillingPageURL)),
		ShowFooter: false,
	})
}

type InvoiceUnpaid struct {
	ToEmail    string
	ToName     string
	OrgName    string
	PaymentURL string
}

// SendInvoiceUnpaid sent after the payment grace period has ended
func (c *Client) SendInvoiceUnpaid(opts *InvoiceUnpaid) error {
	return c.SendCallToAction(&CallToAction{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Invoice for %s is now past due. Org is now hibernated", opts.OrgName),
		PreButton: template.HTML(fmt.Sprintf(`
<b>%s</b> and its projects have been hibernated due to an overdue payment. 
<br /><br />
Restore access by updating your payment information today! 
`, opts.ToName)),
		ButtonText: "Update Payment Info",
		ButtonLink: opts.PaymentURL,
		ShowFooter: true,
	})
}

type SubscriptionCancelled struct {
	ToEmail    string
	ToName     string
	OrgName    string
	PlanName   string
	BillingURL string
	EndDate    time.Time
}

func (c *Client) SendSubscriptionCancelled(opts *SubscriptionCancelled) error {
	return c.SendCallToAction(&CallToAction{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("%s for %s is canceled. Access available until %s", opts.PlanName, opts.OrgName, opts.EndDate.Format(dateFormat)),
		PreButton: template.HTML(fmt.Sprintf(`
We’re sorry to see you go!
<br /><br />
You’ve successfully canceled the %s plan for <b>%s</b>. You’ll still have access to Rill Cloud until <b>%s</b>. After this date, your subscription will expire, and you will no longer have access.
<br /><br />
If you change your mind, you can always reactivate your subscription!
`, opts.PlanName, opts.ToName, opts.EndDate.Format(dateFormat))),
		ButtonText: "Billing Settings",
		ButtonLink: opts.BillingURL,
		PostButton: `If you found that our service did not meet your needs, please contact us via <a href="mailto:support@rilldata.com" style="color:#4736F5">email</a>, or via chat on <a href="https://docs.rilldata.com/contact#in-app-chat" style="color:#4736F5">Rill Developer or Rill Cloud.</a> and we'll do our best to address your feedback and concerns.`,
		ShowFooter: false,
	})
}

type SubscriptionEnded struct {
	ToEmail    string
	ToName     string
	OrgName    string
	BillingURL string
}

func (c *Client) SendSubscriptionEnded(opts *SubscriptionEnded) error {
	return c.SendCallToAction(&CallToAction{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Subscription for %s has now ended. Org is hibernated", opts.OrgName),
		PreButton: template.HTML(fmt.Sprintf(`
Your cancelled subscription for <b>%s</b> has ended and its projects are now <a href="https://docs.rilldata.com/home/FAQ#what-is-project-hibernation">hibernating</a>. We hope you enjoyed using Rill Cloud during your time with us.
<br /><br />
If you’d like to reactivate your subscription and regain access, you can easily do so at any time by renewing your subscription from here:
`, opts.OrgName)),
		ButtonText: "Billing Settings",
		ButtonLink: opts.BillingURL,
		PostButton: `
If you have any feedback about your experience or how we can improve, please feel free to contact us via <a href="mailto:support@rilldata.com" style="color:#4736F5">email</a>, or via chat on <a href="https://docs.rilldata.com/contact#in-app-chat" style="color:#4736F5">Rill Developer or Rill Cloud.</a>
<br /><br />
Thank you for trying Rill Cloud. We hope to see you again in the future!
`,
		ShowFooter: false,
	})
}

type TrialStarted struct {
	ToEmail      string
	ToName       string
	OrgName      string
	FrontendURL  string
	TrialEndDate time.Time
}

func (c *Client) SendTrialStarted(opts *TrialStarted) error {
	return c.SendWelcomeToTrial(&Welcome{
		ToEmail:     opts.ToEmail,
		ToName:      opts.ToName,
		Subject:     fmt.Sprintf("A 30-day free trial for %s has started", opts.OrgName),
		FrontendURL: opts.FrontendURL,
		WelcomeText: template.HTML(fmt.Sprintf(`
You now have access to Rill Cloud until <b>%s</b> to explore all features. Let us know if you need any help along the way!
`, opts.TrialEndDate.Format(dateFormat))),
	})
}

type TrialEndingSoon struct {
	ToEmail      string
	ToName       string
	OrgName      string
	UpgradeURL   string
	TrialEndDate time.Time
}

func (c *Client) SendTrialEndingSoon(opts *TrialEndingSoon) error {
	diff := time.Until(opts.TrialEndDate)
	days := int(math.Round(diff.Hours() / 24))
	return c.SendCallToAction(&CallToAction{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Your Rill Cloud trial for %s is expiring in %d days", opts.OrgName, days),
		PreButton: template.HTML(fmt.Sprintf(`
Your trial for <b>%s</b> ends on <b>%s</b>.
<br /><br />
How's Rill working out for you? Have you checked out our newest features highlighted in our <a href="https://docs.rilldata.com/notes">Release Notes</a>? 
<br /><br />
Our team is here to help you in any way we can, so don't hesitate to contact us via <a href="mailto:support@rilldata.com" style="color:#4736F5">email</a>, or via chat on <a href="https://docs.rilldata.com/contact#in-app-chat" style="color:#4736F5">Rill Developer or Rill Cloud.</a> if you have a question, encounter an issue, or need guidance.
<br /><br />
If you're ready to upgrade, simply click the button below.
`, opts.ToName, opts.TrialEndDate.Format(dateFormat))),
		ButtonText: "Upgrade Now",
		ButtonLink: opts.UpgradeURL,
		ShowFooter: false,
	})
}

type TrialEnded struct {
	ToEmail            string
	ToName             string
	OrgName            string
	UpgradeURL         string
	GracePeriodEndDate time.Time
}

func (c *Client) SendTrialEnded(opts *TrialEnded) error {
	return c.SendCallToAction(&CallToAction{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Your Rill Cloud trial for %s has expired", opts.OrgName),
		PreButton: template.HTML(fmt.Sprintf(`
Hi %s, 
<br /><br />
Your Rill Cloud trial has now expired. <b>%s</b> will be hibernated on <b>%s</b>. We hope you’ve enjoyed using our software. If you’d like to keep using Rill Cloud, upgrade to our Team Plan!
`, opts.ToName, opts.OrgName, opts.GracePeriodEndDate.Format(dateFormat))),
		ButtonText: "Upgrade to Team Plan",
		ButtonLink: opts.UpgradeURL,
		ShowFooter: true,
	})
}

type TrialGracePeriodEnded struct {
	ToEmail    string
	ToName     string
	OrgName    string
	UpgradeURL string
}

func (c *Client) SendTrialGracePeriodEnded(opts *TrialGracePeriodEnded) error {
	return c.SendCallToAction(&CallToAction{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Trial plan grace period for %s has ended. Org is now hibernated", opts.OrgName),
		PreButton: template.HTML(fmt.Sprintf(`
<b>%s</b> and its projects are now <a href="https://docs.rilldata.com/home/FAQ#what-is-project-hibernation">hibernating</a>.
<br /><br />
Reactivate your org by upgrading to the Team Plan today!
`, opts.OrgName)),
		ButtonText: "Upgrade to Team Plan",
		ButtonLink: opts.UpgradeURL,
		PostButton: `
We'd love to hear from you! If you have any feedback about your experience or how we can improve, please feel free to contact us via <a href="mailto:support@rilldata.com" style="color:#4736F5">email</a>, or via chat on <a href="https://docs.rilldata.com/contact#in-app-chat" style="color:#4736F5">Rill Developer or Rill Cloud.</a>
<br /><br />
Thank you for trying Rill Cloud. We hope to see you again in the future!
`,
		ShowFooter: false,
	})
}

type TrialExtended struct {
	ToEmail      string
	ToName       string
	OrgName      string
	TrialEndDate time.Time
}

func (c *Client) SendTrialExtended(opts *TrialExtended) error {
	return c.SendInformational(&Informational{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("Your trial for %s has been extended", opts.OrgName),
		Body:       template.HTML(fmt.Sprintf("Your trial for <b>%q</b> has been extended until %s.", opts.OrgName, opts.TrialEndDate.Format(dateFormat))),
		ShowFooter: true,
	})
}

type PlanUpdate struct {
	ToEmail  string
	ToName   string
	OrgName  string
	PlanName string
}

func (c *Client) SendPlanUpdate(opts *PlanUpdate) error {
	return c.SendInformational(&Informational{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("Your plan for %s has been updated to %s plan", opts.OrgName, opts.PlanName),
		Body:       template.HTML(fmt.Sprintf("<b>%q</b> has been updated to %q plan.", opts.OrgName, opts.PlanName)),
		ShowFooter: true,
	})
}

type SubscriptionRenewed struct {
	ToEmail  string
	ToName   string
	OrgName  string
	PlanName string
}

func (c *Client) SendSubscriptionRenewed(opts *SubscriptionRenewed) error {
	return c.SendInformational(&Informational{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("Your %s subscription for %s plan has been renewed", opts.PlanName, opts.OrgName),
		Body:       template.HTML(fmt.Sprintf("Your subscription for <b>%q</b> has been renewed for %q plan.", opts.OrgName, opts.PlanName)),
		ShowFooter: true,
	})
}

type TeamPlan struct {
	ToEmail          string
	ToName           string
	OrgName          string
	FrontendURL      string
	PlanName         string
	BillingStartDate time.Time
}

// SendTeamPlanStarted sends customised plan started email for Team Plan
func (c *Client) SendTeamPlanStarted(opts *TeamPlan) error {
	return c.SendWelcomeToTeam(&Welcome{
		ToEmail:     opts.ToEmail,
		ToName:      opts.ToName,
		Subject:     fmt.Sprintf("Welcome to the %s plan", opts.PlanName),
		FrontendURL: opts.FrontendURL,
		WelcomeText: template.HTML(fmt.Sprintf(`
Thank you! You’ve successfully upgraded %s to the %s plan.
<br /><br />
Your next billing cycle starts on <b>%s</b>.
`, opts.OrgName, opts.PlanName, opts.BillingStartDate.Format(dateFormat))),
	})
}

// SendTeamPlanRenewal sends customised plan renewed email for Team Plan
func (c *Client) SendTeamPlanRenewal(opts *TeamPlan) error {
	return c.SendWelcomeToTeam(&Welcome{
		ToEmail:     opts.ToEmail,
		ToName:      opts.ToName,
		Subject:     fmt.Sprintf("Your %s plan subscription for %s has been renewed", opts.PlanName, opts.OrgName),
		FrontendURL: opts.FrontendURL,
		WelcomeText: template.HTML(fmt.Sprintf(`
Thank you! You’ve successfully renewed to the <b>%s</b> for %s plan.
<br /><br />
Your next billing cycle starts on %s.
`, opts.OrgName, opts.PlanName, opts.BillingStartDate.Format(dateFormat))),
	})
}
