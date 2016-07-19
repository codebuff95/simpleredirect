package welcome

import (
	"html/template"
	"log"
	"net/http"
	"simpleredirect/user"
)

//WelcomeHandler handles requests made to URL: "/welcome"
//If UserSID is valid, no need to show welcome page, redirect to homepage.
func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	userIsAuthentic := user.Authenticate(r)
	if userIsAuthentic != "" {
		//redirect to homepage.
		log.Println("Request user session is authentic. Redirecting to homepage.")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	//usersid not authentic. Proceed to showing welcome page.
	log.Println("Authentication of UserSID failed. Proceeding to showing welcome page.")
	t, err := template.ParseFiles("simpleredirecttmp/welcome.html")
	if err != nil {
		w.Write([]byte("Error processing requested page. Please try again in some time."))
		log.Println("Error processing home page to client", r.RemoteAddr)
		log.Println(err)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		w.Write([]byte("Error processing requested page. Please try again in some time."))
		log.Println("Error executing home page to client", r.RemoteAddr)
		log.Println(err)
		return
	}
}
