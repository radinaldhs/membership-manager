package email

import (
	"context"
	"fmt"
	"html/template"

	"github.com/wneessen/go-mail"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/errors"
)

type Message struct {
	From         string
	To           []string
	Subject      string
	TemplateName string
	TemplateData any
}

var (
	ErrTemplateNotFound = errors.New(customerror.ErrGroupDataNotFoundErr, "", "email template not found")
)

type EmailServiceProvider interface {
	Send(ctx context.Context, msg Message) error
}

type EmailService struct {
	cfg      config.Email
	smtpCl   *mail.Client
	tmplRoot *template.Template
}

func NewEmailService(emailCfg config.Email) (*EmailService, error) {
	smtpCl, err := mail.NewClient(
		emailCfg.SMTPHost,
		mail.WithPort(emailCfg.SMTPPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(emailCfg.SMTPUsername),
		mail.WithPassword(emailCfg.SMTPPassword),
	)

	if err != nil {
		return nil, err
	}

	tmpl, err := template.ParseGlob(fmt.Sprintf("%s/*.html", emailCfg.TemplateDir))
	if err != nil {
		return nil, err
	}

	return &EmailService{
		cfg:      emailCfg,
		smtpCl:   smtpCl,
		tmplRoot: tmpl,
	}, nil
}

func (svc *EmailService) Send(ctx context.Context, msg Message) error {
	if msg.From == "" {
		msg.From = svc.cfg.Sender
	}

	mailMsg := mail.NewMsg()
	if err := mailMsg.From(msg.From); err != nil {
		return errors.New(customerror.ErrGroupClientErr, "", "error on set sender to email message").WithError(err)
	}

	for _, to := range msg.To {
		if err := mailMsg.AddTo(to); err != nil {
			return errors.New(customerror.ErrGroupClientErr, "", "error on adding recipient to email message").WithError(err)
		}
	}

	mailMsg.Subject(msg.Subject)

	tmpl := svc.tmplRoot.Lookup(msg.TemplateName)
	if tmpl == nil {
		return ErrTemplateNotFound
	}

	if err := mailMsg.SetBodyHTMLTemplate(tmpl, msg.TemplateData); err != nil {
		return customerror.ErrInternal.WithError(err).WithSource()
	}

	if err := svc.smtpCl.DialAndSend(mailMsg); err != nil {
		return customerror.ErrInternal.WithError(err).WithSource()
	}

	return nil
}
