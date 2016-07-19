package redirect

import (
	"html"
	"html/template"
	"log"
	"net/http"
	"simpleredirect/databases"
	"simpleredirect/form"
	"simpleredirect/user"
	"time"
)

type Redirect struct {
	Targetlink string
	Redirectid string
	Createdon  string
}

type RedirectError int

const (
	OKAY             = -1
	MISCELLANEOUS    = 0
	BADREDIRECTID    = 1
	BADTARGETLINK    = 2
	REDIRECTIDEXISTS = 3
)

func (re RedirectError) Error() string {
	if re == MISCELLANEOUS {
		return "MISCELLANEOUS ERROR"
	}
	if re == BADREDIRECTID {
		return "BAD REDIRECT ID"
	}
	if re == BADTARGETLINK {
		return "BAD TARGET LINK"
	}
	if re == REDIRECTIDEXISTS {
		return "REDIRECT ID ALREADY EXISTS"
	}
	return ""
}

func GetRedirect(targetRedirectId string) *Redirect {
	targetRedirect := Redirect{}
	var escapedTarget string
	log.Println("Getting requested RedirectId", targetRedirectId)
	row := databases.GlobalDBM["mydb"].Con.QueryRow("SELECT redirectid,targetlink,createdon FROM redirects WHERE redirectid = '" + targetRedirectId + "'")
	err := row.Scan(&targetRedirect.Redirectid, &escapedTarget, &targetRedirect.Createdon)
	if err != nil {
		log.Println(err)
		log.Println("Returning nil Redirect.")
		return nil
	}
	targetRedirect.Targetlink = html.UnescapeString(escapedTarget)
	log.Println("Success getting target redirectid", targetRedirectId, ", Returning Redirect.")
	return &targetRedirect
}

func GetRedirects(targetUserId string) *[]*Redirect {
	log.Println("Getting active redirects of targetuserid:", targetUserId)
	rows, err := databases.GlobalDBM["mydb"].Con.Query("SELECT redirectid FROM redirects WHERE userid = '" + targetUserId + "'")
	if err != nil {
		log.Println(err)
		return nil
	}
	mySlice := make([]*Redirect, 0)
	for rows.Next() {
		var myRedirect Redirect
		var myRedirectId string
		err = rows.Scan(&myRedirectId)
		if err != nil {
			log.Println(err)
			break
		}
		myRedirect = *GetRedirect(myRedirectId)
		mySlice = append(mySlice, &myRedirect)
	}
	log.Println("Returning active redirects slice of size", len(mySlice))
	if len(mySlice) == 0 {
		return nil
	}
	return &mySlice
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Processing Redirect.")
	requestedRedirectId := template.HTMLEscapeString(r.URL.Path[len("/"):])
	myRedirect := GetRedirect(requestedRedirectId)
	if myRedirect == nil {
		log.Println("Bad redirect request.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	//Valid redirect. Continue to redirecting.
	http.Redirect(w, r, myRedirect.Targetlink, http.StatusSeeOther)
	return
}

func isNotAlphaNumeric(r rune) bool {
	if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return false
	}
	return true
}

func RedirectIdNotGood(redirectid string) bool {
	if len(redirectid) < 3 || len(redirectid) > 16 { //Redirect ID has to be atleast 3 characters
		return true
	}
	for _, r := range redirectid {
		if isNotAlphaNumeric(r) {
			return true
		}
	}
	return false
}

func TargetLinkNotGood(targetlink string) bool {
	if len(targetlink) < 5 || len(targetlink) > 400 {
		return true
	}
	return false
}

func CreateRedirect(r *http.Request, myUserId string) error {
	requestedRedirectId := html.EscapeString(r.Form.Get("requestedredirectid"))
	if RedirectIdNotGood(requestedRedirectId) {
		log.Println("Redirect ID Not Good.")
		return RedirectError(BADREDIRECTID)
	}
	checkRedirect := GetRedirect(requestedRedirectId)
	if checkRedirect != nil { //Redirect with this RedirectId already exists.
		log.Println("Redirect with this RedirectId already exists.")
		return RedirectError(REDIRECTIDEXISTS)
	}
	log.Println("Redirect with this RedirectId already DOES NOT exist.")
	requestedTargetLink := html.EscapeString(r.Form.Get("requestedtargetlink"))
	if TargetLinkNotGood(requestedTargetLink) {
		log.Println("Target Link Not Good.")
		return RedirectError(BADTARGETLINK)
	}
	timeNow := time.Now().Format("2006-01-02 15:04:05")
	stmt, err := databases.GlobalDBM["mydb"].Con.Prepare("INSERT INTO redirects SET userid = '" + myUserId + "', targetlink = '" + requestedTargetLink + "', redirectid = '" + requestedRedirectId + "', createdon = '" + timeNow + "'")
	if err != nil {
		log.Println("Error Preparing Statement for inserting : ", err)
		return RedirectError(MISCELLANEOUS)
	}
	//_,err = stmt.Exec(sm.TableName,sid,rid,expiry)
	_, err = stmt.Exec()
	if err != nil {
		log.Println("Error Executing Statement for inserting :", err)
		return RedirectError(MISCELLANEOUS)
	}
	//Success adding new redirect.
	log.Println("Successfull added new redirect with redirectid:", requestedRedirectId)
	return nil
}

func AddRedirectHandler(w http.ResponseWriter, r *http.Request) {
	userIsAuthentic := user.Authenticate(r)
	if userIsAuthentic == "" {
		//redirect to login page.
		log.Println("Request user session not authentic. Redirecting to login page.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	//usersid authentic. Proceed to processing request.
	log.Println("User session is authentic. Proceeding to processing add redirect request.")
	if r.Method == "GET" {
		log.Println("GET Request not valid on add redirect page. Redirecting to home page.")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	// Method == POST
	r.ParseForm()
	formIsAuthentic := form.Authenticate(r)
	if formIsAuthentic != "" {
		log.Println("Add redirect formSID is authentic")
		myRedirectError := CreateRedirect(r, userIsAuthentic)
		log.Println("Proceeding to parsing addredirect.html")
		t, err := template.ParseFiles("simpleredirecttmp/addredirect.html")
		if err != nil {
			log.Println("Problem parsing addredirect.html.")
			w.Write([]byte("Problem displaying result."))
			return
		}
		t.Execute(w, myRedirectError)
	} else { //FormSid not valid.
		w.Write([]byte("Invalid form submission (Possibly expired). Go back to home page and retry creating redirect."))
		return
	}
}
