package main

import (
	"database/sql"
	"errors"
	"fmt"
	//_ "github.com/mattn/go-sqlite3"
	_ "github.com/go-sql-driver/mysql"
	"log"	
)

var db *sql.DB

func InitDB(driver, dbname string) (*sql.DB, error) {
	var err error
	
	db, err = sql.Open(driver, dbname)	
	_ = syncDB()
	if err != nil {
		return nil, errors.New("Error creating database")
	}
	
	return db, nil
}

func syncDB() error {
	log.Println("--- syncDB")
	//create messages table
	createMessages := `CREATE TABLE messages (
                id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
                uuid char(32) UNIQUE NOT NULL,
                message char(160)   NOT NULL,
                mobile   char(15)    NOT NULL,
                status  INTEGER DEFAULT 0,
                retries INTEGER DEFAULT 0,
                device string NULL,
                created_at TIMESTAMP default CURRENT_TIMESTAMP,
                updated_at TIMESTAMP
            );`
	result, err := db.Exec(createMessages, nil)


	log.Println("--- result",result)
	return err
}

func insertMessage(sms *SMS) error {
	log.Println("--- insertMessage ", sms)
	tx, err := db.Begin()
	if err != nil {
		log.Println("insertMessage: ", err)
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO messages(uuid, message, mobile) VALUES(?, ?, ?)")
	if err != nil {
		log.Println("insertMessage: ", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(sms.UUID, sms.Body, sms.Mobile)
	if err != nil {
		log.Println("insertMessage: ", err)
		return err
	}
	tx.Commit()
	return nil
}

func updateMessageStatus(sms SMS) error {
	log.Println("--- updateMessageStatus ", sms)
	tx, err := db.Begin()
	if err != nil {
		log.Println("updateMessageStatus: ", err)
		return err
	}
	stmt, err := tx.Prepare("UPDATE messages SET status=?, retries=?, device=?, updated_at=NOW() WHERE uuid=?")
	if err != nil {
		log.Println("updateMessageStatus: ", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(sms.Status, sms.Retries, sms.Device, sms.UUID)
	if err != nil {
		log.Println("updateMessageStatus: ", err)
		return err
	}
	tx.Commit()
	return nil
}

func getPendingMessages(bufferSize int) ([]SMS, error) {
	log.Println("--- getPendingMessages ")
	query := fmt.Sprintf("SELECT uuid, message, mobile, status, retries FROM messages WHERE status!=%v AND retries<%v LIMIT %v",
		SMSProcessed, SMSRetryLimit, bufferSize)
	log.Println("getPendingMessages: ", query)

	rows, err := db.Query(query)
	if err != nil {
		log.Println("getPendingMessages: ", err)
		return nil, err
	}
	defer rows.Close()

	var messages []SMS

	for rows.Next() {
		sms := SMS{}
		rows.Scan(&sms.UUID, &sms.Body, &sms.Mobile, &sms.Status, &sms.Retries)
		messages = append(messages, sms)
	}
	rows.Close()
	return messages, nil
}

func GetMessages(filter string) ([]SMS, error) {
	/*
	   expecting filter as empty string or WHERE clauses,
	   simply append it to the query to get desired set out of database
	*/
	log.Println("--- GetMessages")
	query := fmt.Sprintf("SELECT uuid, message, mobile, status, retries, device FROM messages %v", filter)
	log.Println("GetMessages: ", query)

	rows, err := db.Query(query)
	if err != nil {
		log.Println("GetMessages: ", err)
		return nil, err
	}
	defer rows.Close()

	var messages []SMS

	for rows.Next() {
		sms := SMS{}
		rows.Scan(&sms.UUID, &sms.Body, &sms.Mobile, &sms.Status, &sms.Retries, &sms.Device)
		messages = append(messages, sms)
	}
	rows.Close()
	return messages, nil
}

func GetLast7DaysMessageCount() (map[string]int, error) {
	log.Println("--- GetLast7DaysMessageCount")

	rows, err := db.Query(`SELECT DATE_FORMAT( created_at,'%Y-%m-%d') as datestamp,
    COUNT(id) as messagecount FROM messages GROUP BY datestamp
    ORDER BY datestamp DESC LIMIT 7`)
	if err != nil {
		log.Println("GetLast7DaysMessageCount: ", err)
		return nil, err
	}
	defer rows.Close()

	dayCount := make(map[string]int)
	var day string
	var count int
	for rows.Next() {
		rows.Scan(&day, &count)
		dayCount[day] = count
	}
	rows.Close()
	return dayCount, nil
}

func GetStatusSummary() ([]int, error) {
	log.Println("--- GetStatusSummary")

	rows, err := db.Query(`SELECT status, COUNT(id) as messagecount 
    FROM messages GROUP BY status ORDER BY status`)
	if err != nil {
		log.Println("GetStatusSummary: ", err)
		return nil, err
	}
	defer rows.Close()

	var status, count int
	statusSummary := make([]int, 3)
	for rows.Next() {
		rows.Scan(&status, &count)
		statusSummary[status] = count
	}
	rows.Close()
	return statusSummary, nil
}

func deleteMessage(id string) error {
	log.Println("--- deleteMessage ", id)
	tx, err := db.Begin()
	if err != nil {
		log.Println("deleteMessage: ", err)
		return err
	}
	
	stmt, err := tx.Prepare("DELETE FROM messages WHERE id=?")

	if err != nil {
		log.Println("deleteMessage: ", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		log.Println("deleteMessage: ", err)
		return err
	}
	tx.Commit()
	return nil
}

func getMessage(id string) (SMS, error) {
	log.Println("--- getMessage ")
	query := fmt.Sprintf("SELECT uuid,message,mobile,status,retries FROM messages WHERE id=%v",id)
	log.Println("getMessage query: ", query)
	sms := SMS{}
	err := db.QueryRow(query).Scan(&sms.UUID, &sms.Body, &sms.Mobile, &sms.Status, &sms.Retries)
	log.Println("getMessage: ", sms)
	if err != nil {
		log.Println("getMessage: ", err)
		return sms, err
	}	
	return sms, nil
}

func getLastSMS() (int,string) {
	
	query := fmt.Sprintf("SELECT id,phone FROM sms WHERE status=0 and retries<%v ORDER BY id",SMSRetryLimit)
	
	var phone string
	var id int
	err := db.QueryRow(query).Scan(&id,&phone)
	log.Println("getLastSMS Phone: ", phone)
	log.Println("getLastSMS id: ", id)
	if err != nil {
		log.Println("getLastSMS: ", err)
		return id,phone
	}	
	return id,phone
}

func updateSMSRetries(id int) error {
	log.Println("--- updateSMSRetries ", id)
	tx, err := db.Begin()
	if err != nil {
		log.Println("updateSMSRetries: ", err)
		return err
	}
	stmt, err := tx.Prepare("UPDATE sms SET retries=retries+1 WHERE id=?")
	if err != nil {
		log.Println("updateSMSRetries: ", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		log.Println("updateSMSRetries: ", err)
		return err
	}
	tx.Commit()
	return nil
}

func updateSMSSent(id int) error {
	log.Println("--- updateSMSSent ", id)
	tx, err := db.Begin()
	if err != nil {
		log.Println("updateSMSSent: ", err)
		return err
	}
	stmt, err := tx.Prepare("UPDATE sms SET total=total+1, status=1, last_send_date=NOW() WHERE id=?")
	if err != nil {
		log.Println("updateSMSSent: ", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		log.Println("updateSMSSent: ", err)
		return err
	}
	tx.Commit()
	return nil
}
