package email

import (
	"log"
	"net/smtp"
	"strconv"
	"text/template"
)

var EmailTemplate *template.Template
var GlobalEM *EmailManager

type EmailManager struct {
	Username    string
	Password    string
	EmailServer string
	Port        int
	auth        smtp.Auth
}

func InitGlobalEM() {
	log.Println("Initialising Global Email Manager")
	//Example entries for Gmail. Change values accordingly for other mail services.
	GlobalEM = &EmailManager{Username: "yourGmailUsername", Password: "yourGmailPassword", EmailServer: "smtp.gmail.com", Port: 587}
	GlobalEM.auth = smtp.PlainAuth("",
		GlobalEM.Username,
		GlobalEM.Password,
		GlobalEM.EmailServer,
	)
}
func InitWelcomeEmailTemplate() error {
	log.Println("Initialising Welcome Email Template")
	var err error
	EmailTemplate, err = template.ParseFiles("simpleredirecttmp/welcomemail")
	return err
}

func (myem *EmailManager) SendMyEmail(doc []byte, destinationEMail string) error {
	log.Println("Sending Mail to destination", destinationEMail)
	err := smtp.SendMail(myem.EmailServer+":"+strconv.Itoa(myem.Port),
		myem.auth,
		myem.Username,
		[]string{destinationEMail},
		doc)
	return err
}
