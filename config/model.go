package config

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/veonik/squircy2/data"
)

type configRepository struct {
	database *data.DB

	logger *log.Logger
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

		if err := json.Unmarshal(doc, config); err != nil {
			repo.logger.Warnln("error unmarshaling json config:", err)
		}
		config.ID = id

		return
	})
}

func (repo *configRepository) save(config *Configuration) {
	col := repo.database.Use("Settings")
	d := map[string]interface{}{}
	col.ForEachDoc(func(id int, doc []byte) bool {
		if err := json.Unmarshal(doc, &d); err != nil {
			repo.logger.Warnln("error unmarshaling json config:", err)
		}
		return false
	})
	for k, v := range flattenConfig(config) {
		d[k] = v
	}
	if config.ID <= 0 {
		id, err := col.Insert(d)
		if err != nil {
			repo.logger.Warnln("error inserting json config:", err)
		}
		config.ID = id

	} else {
		if err := col.Update(config.ID, d); err != nil {
			repo.logger.Warnln("error updating json config:", err)
		}
	}
}
