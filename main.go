package main

import (
	"fmt"
	"log"	
	"os"	
	"strconv"
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
	
	_numDevices, _ := appConfig.Get("SETTINGS", "DEVICES")
	numDevices, _ := strconv.Atoi(_numDevices)
	log.Println("main: number of devices: ", numDevices)

	var modems []*GSMModem
	for i := 0; i < numDevices; i++ {
		dev := fmt.Sprintf("DEVICE%v", i)
		_port, _ := appConfig.Get(dev, "COMPORT")
		_baud := 115200 //appConfig.Get(dev, "BAUDRATE")
		_devid, _ := appConfig.Get(dev, "DEVID")
		m := &GSMModem{Port: _port, Baud: _baud, Devid: _devid}
		modems = append(modems, m)
	}

	_bufferSize, _ := appConfig.Get("SETTINGS", "BUFFERSIZE")
	bufferSize, _ := strconv.Atoi(_bufferSize)

	_bufferLow, _ := appConfig.Get("SETTINGS", "BUFFERLOW")
	bufferLow, _ := strconv.Atoi(_bufferLow)

	_loaderTimeout, _ := appConfig.Get("SETTINGS", "MSGTIMEOUT")
	loaderTimeout, _ := strconv.Atoi(_loaderTimeout)

	_loaderCountout, _ := appConfig.Get("SETTINGS", "MSGCOUNTOUT")
	loaderCountout, _ := strconv.Atoi(_loaderCountout)

	_loaderTimeoutLong, _ := appConfig.Get("SETTINGS", "MSGTIMEOUTLONG")
	loaderTimeoutLong, _ := strconv.Atoi(_loaderTimeoutLong)

	log.Println("main: Initializing worker")
	InitWorker(modems, bufferSize, bufferLow, loaderTimeout, loaderCountout, loaderTimeoutLong)

	log.Println("main: Initializing server")
	serverhost, _ := appConfig.Get("SETTINGS", "SERVERHOST")
	serverport, _ := appConfig.Get("SETTINGS", "SERVERPORT")
	err = InitServer(serverhost, serverport)
	if err != nil {
		log.Println("main: ", "Error starting server: ", err.Error(), " Aborting")
		os.Exit(1)
	}
	
}
