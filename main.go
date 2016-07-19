package main

import (
	"log"
	"math/rand"
	"net/http"
	"simpleredirect/databases"
	"simpleredirect/email"
	"simpleredirect/handler/home"
	"simpleredirect/handler/login"
	"simpleredirect/handler/register"
	"simpleredirect/handler/welcome"
	"simpleredirect/redirect"
	"simpleredirect/sessions"
	"simpleredirect/user"
	"time"
)

func main() {
	var err error
	rand.Seed(time.Now().UTC().UnixNano())
	databases.InitGlobalDBM()
	sessions.InitGlobalSM()
	databases.GlobalDBM["mydb"] = &databases.DBManager{Name: "mysql", Database: "redirect", User: "root", Password: "123456"}
	err = databases.GlobalDBM["mydb"].Open()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully Opened mydb!")
	}
	//Seperate goroutins for each Table Cleaner.
	go databases.GlobalDBM["mydb"].CleanTable("formsession")
	go databases.GlobalDBM["mydb"].CleanTable("usersession")
	//Initialise User Session Manager in sessions.(GLobal Session Managers Map) using mydb Database and HARDCODED Tablename.
	sessions.GlobalSM["usersm"] = &sessions.SessionManager{Db: databases.GlobalDBM["mydb"], TableName: "usersession"}
	sessions.GlobalSM["formsm"] = &sessions.SessionManager{Db: databases.GlobalDBM["mydb"], TableName: "formsession"}

	//Initialise Global Email Manager.
	email.InitGlobalEM()
	//Initialise Welcome Email Template.
	err = email.InitWelcomeEmailTemplate()
	if err != nil {
		log.Fatal("Could not initialise Welcome Email Template.")
	}
	/*
		http.HandleFunc("/welcome", welcome.WelcomeHandler)
		http.HandleFunc("/r/", redirect.RedirectHandler)
		http.HandleFunc("/addredirect", redirect.AddRedirectHandler)
		http.HandleFunc("/home", home.HomeHandler)
		http.HandleFunc("/login", login.LoginHandler)
		http.HandleFunc("/logout", user.LogoutHandler)
		http.HandleFunc("/register", register.RegisterHandler)
	*/
	http.HandleFunc("/", MyHandler)
	http.ListenAndServe(":8080", nil)
}

func MyHandler(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path
	if requestPath == "/" {
		log.Println("###Welcome Handler###")
		welcome.WelcomeHandler(w, r)
		return
	}
	if requestPath == "/addredirect" {
		log.Println("###Add Redirect Handler###")
		redirect.AddRedirectHandler(w, r)
		return
	}
	if requestPath == "/home" {
		log.Println("###Home Handler###")
		home.HomeHandler(w, r)
		return
	}
	if requestPath == "/login" {
		log.Println("###Login Handler###")
		login.LoginHandler(w, r)
		return
	}
	if requestPath == "/logout" {
		log.Println("###Logout Handler###")
		user.LogoutHandler(w, r)
		return
	}
	if requestPath == "/register" {
		log.Println("###Register Handler###")
		register.RegisterHandler(w, r)
		return
	}
	log.Println("###Redirect Handler###")
	redirect.RedirectHandler(w, r)
	return
}
