package squircy

import (
	"github.com/HouzuoGuo/tiedot/db"
)

func newDatabaseConnection(config *Configuration) (database *db.DB) {
	dir := config.RootPath + "/data"
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
}

func loadConfig(database *db.DB, config *Configuration) {
	repo := configRepository{database}
	repo.FetchInto(config)
	saveConfig(database, config)
}

func saveConfig(database *db.DB, config *Configuration) {
	repo := configRepository{database}
	repo.Save(config)
}
