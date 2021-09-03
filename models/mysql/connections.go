package mysql

import (
	"database/sql"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"

	_ "github.com/go-sql-driver/mysql"
)

var dbConn *sql.DB

// Init initializes the MySQL connection
func Init() error {
	var err error
	dbConn, err = sql.Open("mysql", configmanager.ConfStore.MySQLUser+":"+configmanager.ConfStore.MySQLPassword+"@/"+configmanager.ConfStore.MySQLDB)
	if err != nil {
		return err
	}
	dbConn.SetMaxOpenConns(100)
	return nil
}
