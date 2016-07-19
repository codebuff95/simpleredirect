package user

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"simpleredirect/databases"
	"simpleredirect/email"
	"simpleredirect/sessions"
	"time"
)

type User struct {
	Userid       string
	Firstname    string
	Lastname     string
	Email        string
	Registeredon string
}

const (
	BADPASSWORD   = 1
	BADEMAIL      = 2
	BADFORMSID    = 3
	BADNAME       = 4
	BADADDATTEMPT = 5
)

type RegisterError int

func (re RegisterError) Error() string {
	if re == (BADNAME) {
		return fmt.Sprintf("Problem with entered name.")
	}
	if re == (BADEMAIL) {
		return fmt.Sprintf("Problem with entered email.")
	}
	if re == RegisterError(BADPASSWORD) {
		return fmt.Sprintf("Problem with entered password.")
	}
	if re == RegisterError(BADFORMSID) {
		return fmt.Sprintf("Problem with submitted form SID.")
	}
	return "Miscellaneous Error."
}

//Authenticate authenticates UserSID field of received request. NOTE: r should be parsed for forms.
// returns requesting UserId (empty string if UserSID is not valid).
func Authenticate(r *http.Request) string {
	log.Println("Authenticating User SID on received request.")
	userCookie, err := r.Cookie("usersid")
	if err != nil { //usersid cookie not set on client.
		log.Println("User SID authentication failed")
		return ""
	}
	userSid := userCookie.Value //Authenticate user by its usersid field.
	return sessions.GlobalSM["usersm"].Authenticate(userSid)
}

//AuthenticateLoginAttempt authenticates "email" and "password" fields in the form in incoming
// request 'r'. If authentic, that is, the entered details exist in the database, Login attempt
// is authentic, and a new User Session is created in the database, and returned. Else,
// an unauthentic login attempt triggers return of non-ACTIVE session.
// NOTE 1: r should be parsed for forms.
// NOTE 2: fetching form entries are secured by 'HTML escaping' special characters.
func AuthenticateLoginAttempt(r *http.Request) *sessions.Session {
	var userid string
	log.Println("Authenticating Login credentials.")
	attemptEmail := template.HTMLEscapeString(r.Form.Get("email"))       //Escape special characters for security.
	attemptPassword := template.HTMLEscapeString(r.Form.Get("password")) //Escape special characters for security.
	log.Println("Attempt email :", attemptEmail, "Attempt Password:", attemptPassword)
	row := databases.GlobalDBM["mydb"].Con.QueryRow("SELECT userid FROM user WHERE email = '" + attemptEmail + "' AND password = '" + attemptPassword + "'")
	err := row.Scan(&userid)
	if err != nil { // User does not exist.
		log.Println("User authentication failed.")
		return &sessions.Session{Status: sessions.DELETED}
	}
	//User exists.
	log.Println("User authentication successful. Creating new Session.")
	return sessions.GlobalSM["usersm"].SetSession(userid, time.Hour*24*3) // Session lives in DB for 3 days.
}

func GetUser(targetuserid string) *User {
	targetUser := User{}
	log.Println("Getting requested userid", targetuserid)
	row := databases.GlobalDBM["mydb"].Con.QueryRow("SELECT CONVERT(userid,CHAR(11)),firstname,lastname,email,registeredon FROM user WHERE userid = '" + targetuserid + "'")
	err := row.Scan(&targetUser.Userid, &targetUser.Firstname, &targetUser.Lastname, &targetUser.Email, &targetUser.Registeredon)
	if err != nil {
		log.Println("Could not find target userid", targetuserid, ", Returning nil User.")
		log.Println(err)
		return nil
	}
	log.Println("Success getting target userid", targetuserid, ", Returning User.")
	return &targetUser
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Logout Handler")
	requestingUserId := Authenticate(r)
	if requestingUserId == "" {
		log.Println("Request Session UserId not authentic. No need to logout. Redirecting to login page.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	log.Println("Request Session UserId is authentic. Logging out.")
	myCookie, err := r.Cookie("usersid")
	if err != nil {
		log.Println("UserSid cookie not set on client. Redirect to loginpage.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	userSid := myCookie.Value
	sessionsDeleted := sessions.GlobalSM["usersm"].DeleteSession(userSid)
	log.Println(sessionsDeleted, "sessions deleted. Redirecting to login page.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
	return
}

func passwordNotGood(entered string) bool {
	if len(entered) < 6 || len(entered) > 15 {
		return true
	}
	return false
}

func emailNotGood(entered string) bool {
	//no email logic yet.
	return false
}

func nameNotGood(entered string) bool {
	//no name logic yet.
	return false
}

func AuthenticateRegisterAttempt(r *http.Request) error {
	enteredPassword := template.HTMLEscapeString(r.Form.Get("password"))
	if passwordNotGood(enteredPassword) {
		return RegisterError(BADPASSWORD)
	}
	enteredEmail := template.HTMLEscapeString(r.Form.Get("email"))
	if emailNotGood(enteredEmail) {
		return RegisterError(BADEMAIL)
	}
	enteredFirstName := template.HTMLEscapeString(r.Form.Get("firstname"))
	enteredLastName := template.HTMLEscapeString(r.Form.Get("lastname"))
	if nameNotGood(enteredFirstName) || nameNotGood(enteredLastName) {
		return RegisterError(BADNAME)
	}
	return nil // Authentic Register Attempt.
}

//AddUser adds User into database, and returns a non-nil error if the user could not be added.
//Else, returns a nil error.
func AddUser(r *http.Request) (*User, error) {
	enteredPassword := template.HTMLEscapeString(r.Form.Get("password"))
	enteredEmail := template.HTMLEscapeString(r.Form.Get("email"))
	enteredFirstName := template.HTMLEscapeString(r.Form.Get("firstname"))
	enteredLastName := template.HTMLEscapeString(r.Form.Get("lastname"))
	stmt, err := databases.GlobalDBM["mydb"].Con.Prepare("INSERT INTO user SET firstname = '" + enteredFirstName + "', lastname = '" + enteredLastName + "', password = '" + enteredPassword + "', email = '" + enteredEmail + "'")
	if err != nil {
		log.Println(err)
		return nil, RegisterError(BADADDATTEMPT)
	}
	_, err = stmt.Exec()
	if err != nil {
		log.Println(err)
		return nil, RegisterError(BADADDATTEMPT)
	}
	return &User{Firstname: enteredFirstName, Lastname: enteredLastName, Email: enteredEmail}, nil
}

func WelcomeEmail(myUser *User) error {
	var doc bytes.Buffer
	err := email.EmailTemplate.Execute(&doc, myUser)
	if err != nil {
		return err
	}
	return email.GlobalEM.SendMyEmail(doc.Bytes(), myUser.Email)
}
