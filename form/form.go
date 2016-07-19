package form

import (
	"log"
	"net/http"
	"simpleredirect/sessions"
)

func Authenticate(r *http.Request) string { //NOTE: r should be parsed for forms.
	log.Println("Authenticating Form SID")
	FormSid := r.Form.Get("formsid") //Authenticate form by its formsid field.
	formId := sessions.GlobalSM["formsm"].Authenticate(FormSid)
	if formId != "" {
		sessions.GlobalSM["formsm"].DeleteSession(FormSid)
	}
	return formId
}
