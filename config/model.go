package config

import (
	"encoding/json"
	"fmt"

	"github.com/veonik/squircy2/data"
)

type configRepository struct {
	database *data.DB
}

func flattenConfig(config *Configuration) map[string]interface{} {
	return map[string]interface{}{
		"Network":        config.Network,
		"TLS":            config.TLS,
		"AutoConnect":    config.AutoConnect,
		"Nick":           config.Nick,
		"Username":       config.Username,
		"SASL":           config.SASL,
		"SASLUsername":   config.SASLUsername,
		"SASLPassword":   config.SASLPassword,
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
		"EnableFileAPI":  config.EnableFileAPI,
		"FileAPIRoot":    config.FileAPIRoot,
		"PluginsEnabled": config.PluginsEnabled,
		"PluginsPath":    config.PluginsPath,
	}
}

func (repo *configRepository) fetchInto(config *Configuration) {
	col := repo.database.Use("Settings")
	col.ForEachDoc(func(id int, doc []byte) (moveOn bool) {
		moveOn = false

		fmt.Println("loaded config", json.Unmarshal(doc, config))
		config.ID = id

		return
	})
}

func (repo *configRepository) save(config *Configuration) {
	col := repo.database.Use("Settings")
	data := map[string]interface{}{}
	col.ForEachDoc(func(id int, doc []byte) bool {
		fmt.Println(json.Unmarshal(doc, &data))
		return false
	})
	for k, v := range flattenConfig(config) {
		data[k] = v
	}
	if config.ID <= 0 {
		id, _ := col.Insert(data)
		config.ID = id

	} else {
		col.Update(config.ID, data)
	}
}
