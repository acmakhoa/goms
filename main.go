package main

import (
	"log"	
	"os"
)

func main() {
	appConfig, err := GetConfig("conf.ini")
	log.Println("appConfig",appConfig);
	if err != nil {
		log.Println("main: ", "Invalid config: ", err.Error(), " Aborting")
		os.Exit(1)
	}
	dbuser, _ := appConfig.Get("SETTINGS", "DBUSER")
	dbpass, _ := appConfig.Get("SETTINGS", "DBPASS")
	dbname, _ := appConfig.Get("SETTINGS", "DBNAME")
	dbconnection :=dbuser+":"+dbpass+"@/"+ dbname;
	db, err := InitDB("mysql", dbconnection)

	if err != nil {
		log.Println("main: ", "Error initializing database: ", err, " Aborting")
		os.Exit(1)
	}
	defer db.Close()
	log.Println("main: Initializing worker")
	InitWorker(5,5,5,5,5)

	log.Println("main: Initializing server")
	InitServer("localhost", "8080")
	
}
