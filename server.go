package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

//reposne structure to /sms
type SMSResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

//response structure to /smsdata/
type SMSDataResponse struct {
	Status   int            `json:"status"`
	Message  string         `json:"message"`
	Summary  []int          `json:"summary"`
	DayCount map[string]int `json:"daycount"`
	Messages []SMS          `json:"messages"`
}

// Cache templates
var templates = template.Must(template.ParseFiles("./templates/index.html"))

/* dashboard handlers */

// dashboard
func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("--- indexHandler")
	// templates.ExecuteTemplate(w, "index.html", nil)
	// Use during development to avoid having to restart server
	// after every change in HTML
	t, _ := template.ParseFiles("./templates/index.html")
	t.Execute(w, nil)
}

// handle all static files based on specified path
// for now its /assets
func handleStatic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	static := vars["path"]
	http.ServeFile(w, r, filepath.Join("./assets", static))
}


func InitServer(host string, port string) error {
	log.Println("--- InitServer ", host, port)

	r := mux.NewRouter()
	r.StrictSlash(true)

	r.HandleFunc("/", indexHandler)

	// handle static files
	r.HandleFunc(`/assets/{path:[a-zA-Z0-9=\-\/\.\_]+}`, handleStatic)

	// all API handlers
	api := r.PathPrefix("/api").Subrouter()	
	api.Methods("GET").Path("/logs/").HandlerFunc(getLogsHandler)
	api.Methods("GET").Path("/sms/delete/{id:[0-9]+}").HandlerFunc(deleteSMSHandler)
	api.Methods("GET").Path("/sms/resend/{id:[0-9]+}").HandlerFunc(resendSMSHandler)
	api.Methods("POST").Path("/sms/").HandlerFunc(sendSMSHandler)

	http.Handle("/", r)

	bind := fmt.Sprintf("%s:%s", host, port)
	log.Println("listening on: ", bind)
	return http.ListenAndServe(bind, nil)

}
