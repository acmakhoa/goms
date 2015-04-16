package main
import (
	"encoding/json"
	"strings"
	//"strconv"
	"github.com/satori/go.uuid"	
	"github.com/gorilla/mux"
	"log"
	"net/http"	
)

/* end dashboard handlers */

/* API handlers */

// push sms, allowed methods: POST
func sendSMSHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("--- sendSMSHandler")
	w.Header().Set("Content-type", "application/json")

	r.ParseForm()
	message := r.FormValue("message")
	strMobile := r.FormValue("mobile")
	mobiles  :=strings.Split(strMobile, "\n");
	log.Println("===mobile: ",len(mobiles));
	for i:=0;i<len(mobiles);i++{
		mobile :="+84"+mobiles[i];
		log.Println("===mobile: ",mobile);
		
		uuid := uuid.NewV1()
		sms := &SMS{UUID: uuid.String(), Mobile: mobile, Body: message, Retries: 0}
		EnqueueMessage(sms, true)
	}


	smsresp := SMSResponse{Status: 200, Message: "ok"}
	var toWrite []byte
	toWrite, err := json.Marshal(smsresp)
	if err != nil {
		log.Println(err)
		//lets just depend on the server to raise 500
	}
	w.Write(toWrite)
}
// dumps JSON data, used by log view. Methods allowed: GET
func getLogsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("--- getLogsHandler")
	messages, _ := GetMessages("")
	summary, _ := GetStatusSummary()
	dayCount, _ := GetLast7DaysMessageCount()
	logs := SMSDataResponse{
		Status:   200,
		Message:  "ok",
		Summary:  summary,
		DayCount: dayCount,
		Messages: messages,
	}
	var toWrite []byte
	toWrite, err := json.Marshal(logs)
	if err != nil {
		log.Println(err)
		//lets just depend on the server to raise 500
	}
	w.Header().Set("Content-type", "application/json")
	w.Write(toWrite)
}

// delete sms, allowed methods: POST
func deleteSMSHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("--- deleteSMSHandler")
	w.Header().Set("Content-type", "application/json")
	
	params := mux.Vars(r)
	
	deleteMessage(params["id"])
	log.Println("--- deleteSMSHandler",params["id"])
	
	smsresp := SMSResponse{Status: 200, Message: "ok"}
	var toWrite []byte
	toWrite, err := json.Marshal(smsresp)
	if err != nil {
		log.Println(err)
		//lets just depend on the server to raise 500
	}
	w.Write(toWrite)
}

func resendSMSHandler(w http.ResponseWriter, r *http.Request) {
	
	log.Println("--- resendSMSHandler")
	w.Header().Set("Content-type", "application/json")
	
	params := mux.Vars(r)
	
	message,_ := getMessage(params["id"])
	AddMessage(message)
	
	smsresp := SMSResponse{Status: 200, Message: "ok"}
	var toWrite []byte
	toWrite, err := json.Marshal(smsresp)
	if err != nil {
		log.Println(err)
		//lets just depend on the server to raise 500
	}
	w.Write(toWrite)
}



