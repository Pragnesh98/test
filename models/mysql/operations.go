package mysql

import (
	"strings"
)

// Callback contains callback data
type Callback struct {
	ID          int64  `json:"id"`
	SID         string `json:"sid"`
	CreatedTime string `json:"created_time"`
	UpdatedTime string `json:"updated_time"`
	CallbackURL string `json:"callback_url"`
	Status      string `json:"status"`
	Payload     string `json:"payload"`
}

// InsertSIPBuddies inserts the user into database
func InsertSIPBuddies(user string, secret string) error {
	insertQuery := "INSERT INTO sip_buddies (`name`, `defaultuser`, `md5secret`, `context`, `host`, `nat`, `qualify`, `type`) VALUES (?, ?, ?, 'incoming', 'dynamic', 'yes', 'yes', 'friend')"
	insStmt, err := dbConn.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer insStmt.Close()
	_, err = insStmt.Exec(user, user, secret)
	if err != nil {
		return err
	}
	return nil
}

// SIPBuddyExists fetches the SIP user from the DB
func SIPBuddyExists(name string) (bool, error) {
	var present bool
	selectQuery := "SELECT IF(COUNT(*),'true','false') FROM sip_buddies WHERE name = ?"
	err := dbConn.QueryRow(selectQuery, name).Scan(present)
	if err != nil {
		return false, err
	}
	return present, nil
}

// InsertCallbackRecord inserts the callback record
func InsertCallbackRecord(
	callSID string,
	callbackURL string,
	payload string,
) error {
	insertQuery := "INSERT INTO callbacks (`sid`, `created_time`, `updated_time`, `callback_url`, `status`, `payload`) VALUES (?, NOW(), NOW(), ?, 'scheduled', ?);"
	insStmt, err := dbConn.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer insStmt.Close()
	_, err = insStmt.Exec(callSID, callbackURL, payload)
	if err != nil {
		return err
	}
	return nil
}

// GetScheduledCallbacks gets the pending callbacks
func GetScheduledCallbacks() ([]Callback, []int64, error) {
	selQuery := "SELECT * FROM callbacks WHERE status = 'scheduled' ORDER by created_time DESC LIMIT 20"
	selDB, err := dbConn.Query(selQuery)
	if err != nil {
		return nil, nil, err
	}
	var callbacks []Callback
	var iDs []int64
	for selDB.Next() {
		var cB Callback
		if err = selDB.Scan(&cB.ID, &cB.SID, &cB.CreatedTime, &cB.UpdatedTime, &cB.CallbackURL, &cB.Status, &cB.Payload); err != nil {
			return callbacks, iDs, err
		}
		callbacks = append(callbacks, cB)
		iDs = append(iDs, cB.ID)
	}
	return callbacks, iDs, nil
}

// MarkCallbackScheduled status
func MarkCallbackScheduled(
	ID int64,
	callbackURL string,
) error {
	updateQuery := "UPDATE callbacks SET status='scheduled', updated_time=NOW() WHERE id= ? AND callback_url= ?"
	updateDB, err := dbConn.Prepare(updateQuery)
	if err != nil {
		return err
	}
	defer updateDB.Close()
	_, err = updateDB.Exec(ID, callbackURL)
	if err != nil {
		return err
	}
	return nil
}

// MarkCallbackInProgress status
func MarkCallbackInProgress(
	iDs []int64,
) error {
	args := make([]interface{}, len(iDs))
	for i, id := range iDs {
		args[i] = id
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	updateQuery := `UPDATE callbacks SET status='in-progress', updated_time=NOW() WHERE id IN (?` + strings.Repeat(",?", len(args)-1) + `)`
	rows, err := dbConn.Query(updateQuery, args...)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()
	return tx.Commit()
}

// MarkCallbackCompleted status
func MarkCallbackCompleted(
	ID int64,
	callbackURL string,
) error {
	updateQuery := "UPDATE callbacks SET status='completed', updated_time=NOW() WHERE id= ? AND callback_url= ?"
	updateDB, err := dbConn.Prepare(updateQuery)
	if err != nil {
		return err
	}
	defer updateDB.Close()
	_, err = updateDB.Exec(ID, callbackURL)
	if err != nil {
		return err
	}
	return nil
}
