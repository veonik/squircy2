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
		"Network":        config.Network,
		"TLS":            config.TLS,
		"AutoConnect":    config.AutoConnect,
		"Nick":           config.Nick,
		"Username":       config.Username,
		"OwnerNick":      config.OwnerNick,
		"OwnerHost":      config.OwnerHost,
		"WebInterface":   config.WebInterface,
		"HTTPHostPort":   config.HTTPHostPort,
		"HTTPS":          config.HTTPS,
		"RequireHTTPS":   config.RequireHTTPS,
		"SSLHostPort":    config.SSLHostPort,
		"SSLCertFile":    config.SSLCertFile,
		"SSLCertKey":     config.SSLCertKey,
		"HTTPAuth":       config.HTTPAuth,
		"AuthUsername":   config.AuthUsername,
		"AuthPassword":   config.AuthPassword,
		"ScriptsAsFiles": config.ScriptsAsFiles,
		"ScriptsPath":    config.ScriptsPath,
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
