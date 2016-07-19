package register

import (
	"html/template"
	"log"
	"net/http"
	"simpleredirect/form"
	"simpleredirect/sessions"
	"simpleredirect/user"
	"time"
)

//DisplayRegisterPage generates and displays the register form to 'w', replying to request 'r'.
func DisplayRegisterPage(w http.ResponseWriter, r *http.Request) {
	thisSession := sessions.GlobalSM["formsm"].SetSession("0", time.Minute*5) // Form valid for 5 minutes.
	if thisSession == nil {
		//fmt.Fprintf(w,"Error showing login page. Please try again in some time.")
		log.Println("Error creating session for formsid.")
		return
	}
	if thisSession.Status == sessions.ACTIVE {
		t, err := template.ParseFiles("simpleredirecttmp/register.html")
		if err != nil {
			log.Println(err)
		}
		t.Execute(w, thisSession.Sid)
		log.Println("Generated new register form to client", r.RemoteAddr, "with formsid =", thisSession.Sid)
	}
}

//RegisterHandler handles requests made to URL: "/register"
//If UserSID is valid, no need to register, redirect to homepage. Else,
//If method of request is "GET", DisplayRegisterPage is called.Else,
//submitted register form is authenticated.
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Register Handler")
	userIsAuthentic := user.Authenticate(r)
	if userIsAuthentic != "" {
		//redirect to homepage.
		log.Println("Request user session is authentic. Redirecting to homepage.")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	//usersid not authentic. Proceed to processing request.
	if r.Method == "GET" {
		DisplayRegisterPage(w, r)
	} else { // Method == POST
		r.ParseForm()
		formIsAuthentic := form.Authenticate(r)
		if formIsAuthentic != "" { //if form is authentic.
			//r.ParseForm() //ParseForm already called above. No need for redundant calls.
			myRegisterErr := user.AuthenticateRegisterAttempt(r) // returns RegisterError.
			if myRegisterErr != nil {                            //Invalid register attempt.
				log.Println("Invalid Register Attempt.")
				log.Println(myRegisterErr)
				w.Write([]byte("BAD Register Attempt. " + myRegisterErr.Error() + " Try again."))
				return
			} else { //Valid register attempt.
				log.Println("Valid Register Attempt. Adding User to Database")
				myUser, myRegisterErr := user.AddUser(r)
				if myRegisterErr != nil {
					log.Println("Could not add user to database.")
					w.Write([]byte("Could not create account. Please try again."))
					return
				} else {
					log.Println("Success adding user to database.")
					//Send Welcome Email to user.
					err := user.WelcomeEmail(myUser)
					if err != nil {
						log.Println("Could not send Welcome Email to new user", myUser, err)
					} else {
						log.Println("WelcomeEmail sent Successfully to user", myUser)
					}
				}
			}
			t, err := template.ParseFiles("simpleredirecttmp/registerresult.html")
			if err != nil {
				log.Println("Could not parse template registerresult.html. Redirecting to register page.")
				log.Println(err)
				http.Redirect(w, r, "/register", http.StatusSeeOther)
				return
			}
			log.Println("Successfully parsed template registerresult.html. Executing template.")
			t.Execute(w, myRegisterErr)
		} else { //FormSid not valid.
			log.Println("FormSid of register form invalid. Redisplaying register form.")
			DisplayRegisterPage(w, r)
		}
	}
}
