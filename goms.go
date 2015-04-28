package main

import (
	"github.com/haxpax/goserial"
	"log"	
	"strings"
	"time"
)

//TODO: should be configurable
const SMSRetryLimit = 2

type GSMModem struct {
	Port   string
	Baud   int
	Status bool
	Conn   serial.ReadWriteFlushCloser
	Devid  string
}

func (m *GSMModem) Connect() error {
	//log.Println("--- Connect")
	c := &serial.Config{Name: m.Port, Baud: m.Baud, ReadTimeout: time.Second}
	s, err := serial.OpenPort(c)
	if err == nil {
		m.Status = true
		m.Conn = s
	}
	return err
}

func (m *GSMModem) SendCommand(command string, waitForOk bool) string {
	log.Println("--- SendCommand: ", command)
	var status string = ""
	m.Conn.Flush()
	_, err := m.Conn.Write([]byte(command))
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 32)
	var loop int = 1
	if waitForOk {
		loop = 10
	}
	for i := 0; i < loop; i++ {
		// ignoring error as EOF raises error on Linux
		n, _ := m.Conn.Read(buf)
		if n > 0 {
			status = string(buf[:n])
			log.Printf("SendCommand: rcvd %d bytes: %s\n", n, status)
			if strings.HasSuffix(status, "OK\r\n") || strings.HasSuffix(status, "ERROR\r\n") {
				break
			}
		}
	}
	return status
}

func (m *GSMModem) SendSMS(mobile string, message string) int {
	log.Println("--- SendSMS ", mobile, message)

	// Put Modem in SMS Text Mode
	m.SendCommand("AT+CMGF=1\r", false)

	m.SendCommand("AT+CMGS=\""+mobile+"\"\r", false)

	// EOM CTRL-Z = 26
	status := m.SendCommand(message+string(26), true)
	log.Println("--- SendSMS Status ==== ", status)
	if strings.HasSuffix(status, "OK\r\n") {
		return SMSProcessed
	} else if strings.HasSuffix(status, "ERROR\r\n") {
		return SMSError
	} else {
		return SMSPending
	}

}

func (m *GSMModem) ProcessMessages() {
	defer func() {
		log.Println("--- deferring ProcessMessage")
		m.Status = false
	}()

	//log.Println("--- ProcessMessage")
	for {
		message := <-messages
		log.Println("processing: ", message.UUID, m.Devid)

		message.Status = m.SendSMS(message.Mobile, message.Body)
		message.Device = m.Devid
		message.Retries++
		//updateMessageStatus(message)
		log.Println("processing updateSMSSent:  ", message.Id)

		//save status sent sms
		if message.Status == SMSProcessed{
			updateSMSSent(message.Id)
		}

		if message.Status != SMSProcessed && message.Retries < SMSRetryLimit {
			// push message back to queue until either it is sent successfully or
			// retry count is reached
			// I can't push it to channel directly. Doing so may cause the sms to be in
			// the queue twice. I don't want that
			updateSMSRetries(message.Id)
		}
		time.Sleep(3000 * time.Microsecond)
	}
}
