package data

import (
	"github.com/HouzuoGuo/tiedot/db"
)

func NewDatabaseConnection(rootPath string) (database *db.DB) {
	dir := rootPath + "/data"
	database, err := db.OpenDB(dir)
	if err != nil {
		panic(err)
	}

	initDatabase(database)

	return
}

func initDatabase(database *db.DB) {
	col := database.Use("Settings")
	if col == nil {
		err := database.Create("Settings")
		if err != nil {
			panic(err)
		}
	}

	col = database.Use("Scripts")
	if col == nil {
		err := database.Create("Scripts")
		if err != nil {
			panic(err)
		}
	}

	col = database.Use("Webhooks")
	if col == nil {
		err := database.Create("Webhooks")
		if err != nil {
			panic(err)
		}
	}
}
