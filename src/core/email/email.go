package email

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"text/template"

	// "github.com/Emyrk/LendingBot/slack"
	log "github.com/sirupsen/logrus"
)

var emailLog = log.WithFields(log.Fields{
	"package": "email",
})

const (
	SMTP_EMAIL_USER     = "hodlzonesite@gmail.com"
	SMTP_EMAIL_PASS     = "cqvrijwdlbxzrawa"
	SMTP_EMAIL_HOST     = "smtp.gmail.com"
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
	// llog := emailLog.WithField("method", "SendEmail")

	auth := smtp.PlainAuth("", SMTP_EMAIL_USER, SMTP_EMAIL_PASS, SMTP_EMAIL_HOST)

	toStr := ""
	for i, e := range r.to {
		toStr += e
		if i < len(r.to)-1 {
			toStr += ","
		}
	}

	header := make(map[string]string)
	header["From"] = r.from
	header["Subject"] = r.subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(r.Body))
	err := smtp.SendMail(
		SMTP_EMAIL_HOST+":"+SMTP_EMAIL_PORT,
		auth,
		r.from,
		r.to,
		[]byte(message),
	)
	return err
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
