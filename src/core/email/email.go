package email

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/smtp"
	"text/template"
)

const (
	SMTP_EMAIL_USER     = "general2@hodl.zone"
	SMTP_EMAIL_PASS     = "aka@>35RNKANSDFKN2#k%rnJKABSDF"
	SMTP_EMAIL_HOST     = "smtp.zoho.com"
	SMTP_EMAIL_PORT     = "587"
	SMTP_EMAIL_NO_REPLY = "no_reply@hodl.zone"
)

//Request struct
type Request struct {
	from    string
	to      []string
	subject string
	Body    string
}

func NewHTMLRequest(from string, to []string, subject string) *Request {
	return &Request{
		from:    from,
		to:      to,
		subject: subject,
	}
}

func (r *Request) SendEmail() error {
	auth := smtp.PlainAuth("", SMTP_EMAIL_USER, SMTP_EMAIL_PASS, SMTP_EMAIL_HOST)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + r.subject + "!\n"
	msg := []byte(subject + mime + "\n" + r.Body)
	addr := SMTP_EMAIL_HOST + ":" + SMTP_EMAIL_PORT

	if err := smtp.SendMail(addr, auth, r.from, r.to, msg); err != nil {
		return err
	}
	return nil
}

func (r *Request) ParseTemplate(file string, data interface{}) error {

	t := template.New("Penis")
	reader, err := Open(file)
	if err != nil {
		return err
	}

	d, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	t, err = t.Parse(string(d))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	err = t.Execute(writer, data)
	if err != nil {
		return err
	}
	writer.Flush()
	r.Body = string(buf.Bytes())

	return nil
}
