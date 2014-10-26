package squircy

import (
	"encoding/json"
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
	col := database.Use("Settings")

	var settings []byte
	var identifier int
	col.ForEachDoc(func(id int, doc []byte) (moveOn bool) {
		moveOn = false

		settings = doc
		identifier = id

		return
	})

	if len(settings) == 0 {
		// Persist existing settings and return
		identifier, err := col.Insert(map[string]interface{}{
			"Network":   config.Network,
			"Nick":      config.Nick,
			"Username":  config.Username,
			"Password":  config.Password,
			"OwnerNick": config.OwnerNick,
			"OwnerHost": config.OwnerHost,
		})

		config.ID = identifier

		if err != nil {
			panic(err)
		}

		return
	}

	json.Unmarshal(settings, config)
	config.ID = identifier
}

func saveConfig(database *db.DB, config *Configuration) {
	col := database.Use("Settings")

	col.Update(config.ID, map[string]interface{}{
		"Network":   config.Network,
		"Nick":      config.Nick,
		"Username":  config.Username,
		"Password":  config.Password,
		"OwnerNick": config.OwnerNick,
		"OwnerHost": config.OwnerHost,
	})
}
