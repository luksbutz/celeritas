package mailer

import (
	"bytes"
	"fmt"
	apimail "github.com/ainsleyclark/go-mail"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"time"
)

type Mail struct {
	Domain      string
	Templates   string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Jobs        chan Message
	Results     chan Result
	API         string
	APIKey      string
	APIUrl      string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Template    string
	Attachments []string
	Data        interface{}
}

type Result struct {
	Success bool
	Error   error
}

func (m *Mail) ListenForMail() {
	for {
		msg := <-m.Jobs
		err := m.Send(msg)
		if err != nil {
			m.Results <- Result{false, err}
		} else {
			m.Results <- Result{true, nil}
		}
	}
}

func (m *Mail) Send(msg Message) error {
	// TODO: are we using an API or SMTP?
	if len(m.API) > 0 && len(m.APIKey) > 0 && len(m.APIUrl) > 0 && m.API != "smtp" {
		return m.ChooseAPI(msg)
	} else {
		return m.SendSMTPMessage(msg)
	}
}

func (m *Mail) ChooseAPI(msg Message) error {
	switch m.API {
	case "mailgun", "sparkpost", "sendgrid":
		return m.SendUsingAPI(msg, m.API)
	default:
		return fmt.Errorf("unknown api %s; only mailgun, sparkpost or sendgrid accepted", m.API)
	}
}

func (m *Mail) SendUsingAPI(msg Message, transport string) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	cfg := apimail.Config{
		URL:         m.APIUrl,
		APIKey:      m.APIKey,
		Domain:      m.Domain,
		FromAddress: msg.From,
		FromName:    msg.FromName,
	}

	driver, err := apimail.NewClient(transport, cfg)
	if err != nil {
		return err
	}

	htmlMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	tx := &apimail.Transmission{
		Recipients: []string{msg.To},
		Subject:    msg.Subject,
		HTML:       htmlMessage,
		PlainText:  plainMessage,
	}

	// add attachments
	err = m.addAPIAttachments(msg, tx)
	if err != nil {
		return err
	}

	_, err = driver.Send(tx)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mail) addAPIAttachments(msg Message, tx *apimail.Transmission) error {
	if len(msg.Attachments) > 0 {
		var attachments apimail.Attachments

		for _, x := range msg.Attachments {
			var attach apimail.Attachment
			content, err := ioutil.ReadFile(x)
			if err != nil {
				return err
			}
			attach.Filename = filepath.Base(x)
			attach.Bytes = content
			attachments = append(attachments, attach)
		}

		tx.Attachments = attachments
	}

	return nil
}

// SendSMTPMessage builds and sends an email message using SMTP. This is called by ListenForMail,
// and can also be called directly when necessary
func (m *Mail) SendSMTPMessage(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}
	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	htmlMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject).
		SetBody(mail.TextHTML, htmlMessage).
		AddAlternative(mail.TextPlain, plainMessage)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	return email.Send(smtpClient)
}

func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSL
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.html.tmpl", m.Templates, msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	return m.inlineCSS(tpl.String())
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.plain.tmpl", m.Templates, msg.Template)

	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer

	if err := t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", nil
	}

	return tpl.String(), nil
}

func (m *Mail) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}
