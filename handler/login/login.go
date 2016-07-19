package login

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"simpleredirect/form"
	"simpleredirect/sessions"
	"simpleredirect/user"
	"time"
)

//DisplayLoginPage generates and displays the login form to 'w', replying to request 'r'.
func DisplayLoginPage(w http.ResponseWriter, r *http.Request) {
	thisSession := sessions.GlobalSM["formsm"].SetSession("0", time.Minute*5) // Form valid for 5 minutes.
	if thisSession == nil {
		fmt.Fprintf(w, "Error showing login page. Please try again in some time.")
	}
	if thisSession.Status == sessions.ACTIVE {
		t, _ := template.ParseFiles("simpleredirecttmp/login.html")
		t.Execute(w, thisSession.Sid)
		log.Println("Generated new form to client", r.RemoteAddr, "with formsid =", thisSession.Sid)
	}
}

//LoginHandler handles requests made to URL: "/login"
//If UserSID is valid, no need to login, redirect to homepage. Else,
//If method of request is "GET", DisplayLoginPage is called.Else,
//submitted login form is authenticated.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	userIsAuthentic := user.Authenticate(r)
	if userIsAuthentic != "" {
		//redirect to homepage.
		log.Println("Request user session is authentic. Redirecting to homepage.")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	//usersid not authentic. Proceed to processing request.
	if r.Method == "GET" {
		DisplayLoginPage(w, r)
	} else { // Method == POST
		r.ParseForm()
		formIsAuthentic := form.Authenticate(r)
		if formIsAuthentic != "" {
			//Authenticate User Begin.
			userSession := user.AuthenticateLoginAttempt(r)
			if userSession == nil || userSession.Status != sessions.ACTIVE { //Invalid login attempt.
				//fmt.Fprintf(w,"Invalid Login attempt. Retry.\n")
				DisplayLoginPage(w, r)
			} else { //Authentic user login. Set usersid Cookie at client.
				userSidCookie := &http.Cookie{Name: "usersid", Value: userSession.Sid}
				expiry := time.Now().Add(time.Hour * 24 * 2) // Cookie is valid till 2 days.
				if r.Form.Get("rememberme") != "" {          //rememberme selected in login form.
					log.Println("Remember me WAS selected in form.")
					userSidCookie.Expires = expiry //userSidCookie is made a Persistent Cookie.
				} else {
					log.Println("Remember me WAS NOT selected in form.")
					//No need to set expiry field in cookie if rememberme field was not clicked.
				}
				http.SetCookie(w, userSidCookie)
				log.Println("usersid Cookie successfully set on client. Redirecting to homepage.")
				http.Redirect(w, r, "/home", http.StatusSeeOther)
				return
			}
			//Authenticate User End.
		} else { //FormSid not valid.
			DisplayLoginPage(w, r)
		}
	}
}
