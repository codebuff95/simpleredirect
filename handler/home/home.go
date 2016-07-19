package home

import (
	"html/template"
	"log"
	"net/http"
	"simpleredirect/redirect"
	"simpleredirect/sessions"
	"simpleredirect/user"
	"time"
)

//HomeProfile type struct facilitates the passing of whole details needed in a homepage, to template.Execute
//function.
type HomeProfile struct {
	RequestingUser  user.User
	FormSid         string
	ActiveRedirects *[]*redirect.Redirect
}

//HomeHandler handles all requests made to URL: "/home".
//It shows home page of user if the userSID is valid, else redirects to login page.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	requestingUserId := user.Authenticate(r)
	if requestingUserId == "" {
		log.Println("Request Session UserId not authentic. Redirecting to login page.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	log.Println("Request Session UserId is authentic. Processing home display request.")
	log.Println("Requesting home of userid:", requestingUserId)
	requestingUser := user.GetUser(requestingUserId)
	if requestingUser == nil {
		//fmt.Fprintf(w,"User does not exist. Redirecting to login...")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	myHomeProfile := HomeProfile{}
	myHomeProfile.RequestingUser = *requestingUser
	myHomeProfile.ActiveRedirects = redirect.GetRedirects(requestingUserId)
	//Create formsid
	thisSession := sessions.GlobalSM["formsm"].SetSession("0", time.Minute*30) // Form valid for 30 minutes.
	if thisSession == nil {
		log.Println("Error creating session for form. Please try again in some time.")
		return
	}
	if thisSession.Status == sessions.ACTIVE {
		myHomeProfile.FormSid = thisSession.Sid
		log.Println("Generating new form to client", r.RemoteAddr, "with formsid =", thisSession.Sid)
	}
	t, err := template.ParseFiles("simpleredirecttmp/home.html")
	if err != nil {
		log.Println("Problem parsing home page to client. ", err)
		return
	}
	t.Execute(w, myHomeProfile)
}
