package config

import (
	"encoding/json"

	"github.com/HouzuoGuo/tiedot/db"
)

type configRepository struct {
	database *db.DB
}

func flattenConfig(config *Configuration) map[string]interface{} {
	return map[string]interface{}{
		"Network":   config.Network,
		"Nick":      config.Nick,
		"Username":  config.Username,
		"OwnerNick": config.OwnerNick,
		"OwnerHost": config.OwnerHost,
		"TLS":       config.TLS,
	}
}

func (repo *configRepository) fetchInto(config *Configuration) {
	col := repo.database.Use("Settings")
	col.ForEachDoc(func(id int, doc []byte) (moveOn bool) {
		moveOn = false

		json.Unmarshal(doc, config)
		config.ID = id

		return
	})
}

func (repo *configRepository) save(config *Configuration) {
	col := repo.database.Use("Settings")
	data := flattenConfig(config)

	if config.ID <= 0 {
		id, _ := col.Insert(data)
		config.ID = id

	} else {
		col.Update(config.ID, data)
	}
}
