package proto

import (
	"fmt"
	"html/template"
	"net/url"
	"time"

	"euphoria.leet.nu/heim/proto/emails"
	"euphoria.leet.nu/heim/proto/jobs"
	"euphoria.leet.nu/heim/proto/snowflake"
	"euphoria.leet.nu/heim/templates"
	"github.com/euphoria-io/scope"
)

const (
	PasswordChangedEmail       = "password-changed"
	PasswordResetEmail         = "password-reset"
	RoomInvitationEmail        = "room-invitation"
	RoomInvitationWelcomeEmail = "room-invitation-welcome"
	VerificationEmail          = "verification"
	WelcomeEmail               = "welcome"
)

type EmailTracker interface {
	Get(ctx scope.Context, accountID snowflake.Snowflake, id string) (*emails.EmailRef, error)
	List(ctx scope.Context, accountID snowflake.Snowflake, n int, before time.Time) ([]*emails.EmailRef, error)
	MarkDelivered(ctx scope.Context, accountID snowflake.Snowflake, id string) error
	Send(
		ctx scope.Context, js jobs.JobService, templater *templates.Templater, deliverer emails.Deliverer,
		account Account, to, templateName string, data interface{}) (*emails.EmailRef, error)
}

type CommonEmailParams struct {
	emails.CommonData

	EmailDomain   string        `yaml:"email_domain"`
	SiteName      string        `yaml:"site_name"`
	SiteURL       string        `yaml:"site_url"`
	HelpAddress   template.HTML `yaml:"help_address"`
	SenderAddress template.HTML `yaml:"sender_address"`
}

func (p *CommonEmailParams) SiteURLShort() template.HTML {
	return template.HTML(p.CommonData.LocalDomain)
}

func (p *CommonEmailParams) EmailPreferencesURL() template.HTML {
	// TODO: incorporate token
	return template.HTML(fmt.Sprintf("%s/prefs/emails", p.SiteURL))
}

type VerificationEmailParams struct {
	CommonEmailParams
	VerificationToken string
}

func (p VerificationEmailParams) Subject() template.HTML {
	return template.HTML("Please verify your email address")
}

func (p VerificationEmailParams) VerifyEmailURL() template.HTML {
	return verificationURL(p.SiteURL, p.AccountEmailAddress, p.VerificationToken)
}

type WelcomeEmailParams struct {
	CommonEmailParams
	VerificationToken string
}

func (p WelcomeEmailParams) Subject() template.HTML {
	return template.HTML(fmt.Sprintf("Welcome to %s!", p.SiteName))
}

func (p *WelcomeEmailParams) VerifyEmailURL() template.HTML {
	return verificationURL(p.SiteURL, p.AccountEmailAddress, p.VerificationToken)
}

type PasswordChangedEmailParams struct {
	CommonEmailParams
	AccountName string
}

func (p PasswordChangedEmailParams) Subject() template.HTML {
	return template.HTML(fmt.Sprintf("Your %s account password has been changed", p.SiteName))
}

type PasswordResetEmailParams struct {
	CommonEmailParams
	AccountName  string
	Confirmation string
}

func (p PasswordResetEmailParams) Subject() template.HTML {
	return template.HTML(fmt.Sprintf("Password reset request for your %s account", p.SiteName))
}

func (p PasswordResetEmailParams) ResetPasswordURL() template.HTML {
	v := url.Values{
		"confirmation": []string{p.Confirmation},
	}
	u := url.URL{
		Path:     "/prefs/reset-password",
		RawQuery: v.Encode(),
	}
	return template.HTML(p.SiteURL + u.String())
}

type RoomInvitationEmailParams struct {
	CommonEmailParams
	AccountName   string
	RoomName      string
	SenderName    string
	SenderMessage string
}

func (p RoomInvitationEmailParams) Subject() template.HTML {
	return template.HTML(fmt.Sprintf("%s invites you to join &%s", p.SenderName, p.RoomName))
}

func (p RoomInvitationEmailParams) RoomURL() template.HTML {
	return template.HTML(fmt.Sprintf("%s/room/%s", p.SiteURL, p.RoomName))
}

type RoomInvitationWelcomeEmailParams struct {
	CommonEmailParams
	AccountName   string
	RoomName      string
	RoomPrivacy   string
	SenderName    string
	SenderMessage string
}

func (p RoomInvitationWelcomeEmailParams) Subject() template.HTML {
	return template.HTML(fmt.Sprintf("%s invites you to join a chatroom on %s", p.SenderName, p.SiteName))
}

func (p RoomInvitationWelcomeEmailParams) RoomURL() template.HTML {
	return template.HTML(fmt.Sprintf("%s/room/%s", p.SiteURL, p.RoomName))
}

var (
	DefaultCommonEmailParams = CommonEmailParams{
		CommonData: emails.CommonData{
			LocalDomain: "heim.invalid",
		},
		SenderAddress: "noreply@heim.invalid",
		HelpAddress:   "help@heim.invalid",
		SiteName:      "heim",
		SiteURL:       "https://heim.invalid",
	}

	EmailScenarios = map[string]map[string]templates.TemplateTest{
		WelcomeEmail: map[string]templates.TemplateTest{
			"default": templates.TemplateTest{
				Data: &WelcomeEmailParams{
					CommonEmailParams: DefaultCommonEmailParams,
					VerificationToken: "token",
				},
			},
		},

		PasswordChangedEmail: map[string]templates.TemplateTest{
			"default": templates.TemplateTest{
				Data: &PasswordChangedEmailParams{
					CommonEmailParams: DefaultCommonEmailParams,
					AccountName:       "yourname",
				},
			},
		},

		PasswordResetEmail: map[string]templates.TemplateTest{
			"default": templates.TemplateTest{
				Data: &PasswordResetEmailParams{
					CommonEmailParams: DefaultCommonEmailParams,
					AccountName:       "yourname",
				},
			},
		},

		RoomInvitationEmail: map[string]templates.TemplateTest{
			"default": templates.TemplateTest{
				Data: &RoomInvitationEmailParams{
					CommonEmailParams: DefaultCommonEmailParams,
					SenderName:        "(‿|‿)",
					RoomName:          "butts",
					SenderMessage:     "hey, i heard you like butts",
				},
			},
		},

		RoomInvitationWelcomeEmail: map[string]templates.TemplateTest{
			"default": templates.TemplateTest{
				Data: &RoomInvitationWelcomeEmailParams{
					CommonEmailParams: DefaultCommonEmailParams,
					SenderName:        "thatguy",
					RoomName:          "cabal",
					RoomPrivacy:       "private",
					SenderMessage:     "let's move our machinations here",
				},
			},
		},
	}
)

func ValidateEmailTemplates(templater *templates.Templater) []error {
	errors := []error{}
	for templateName, testCases := range EmailScenarios {
		testList := make([]templates.TemplateTest, 0, len(testCases))
		for _, testCase := range testCases {
			testList = append(testList, testCase)
		}
		errors = append(errors, templater.Validate(templateName, testList...)...)
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

func verificationURL(siteURL, email, token string) template.HTML {
	v := url.Values{
		"email": []string{email},
		"token": []string{token},
	}
	u := url.URL{
		Path:     "/prefs/verify",
		RawQuery: v.Encode(),
	}
	return template.HTML(siteURL + u.String())
}
