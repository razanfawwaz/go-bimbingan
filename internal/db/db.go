package db

import (
	"database/sql"
	"fmt"

	"log"

	_ "github.com/lib/pq"
	"github.com/razanfawwaz/bimbingan/util"
)

var DB *sql.DB

func Init() {
	dbHost := util.GetConfig("DB_HOST")
	dbPort := util.GetConfig("DB_PORT")
	dbUser := util.GetConfig("DB_USER")
	dbPass := util.GetConfig("DB_PASS")
	dbName := util.GetConfig("DB_NAME")

	var err error
	dbUrl := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)

	DB, err = sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("error when connecting to db: %s", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("error when pinging db: %s", err)
	}
}
