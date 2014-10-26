package squircy

import (
	"encoding/json"
	"github.com/HouzuoGuo/tiedot/db"
)

type scriptType string

const (
	scriptJavascript scriptType = "Javascript"
	scriptLua                   = "Lua"
	scriptLisp                  = "Lisp"
)

type persistentScript struct {
	ID      int
	Type    scriptType
	Title   string
	Body    string
	Enabled bool
}

type scriptRepository struct {
	database *db.DB
}

func hydrateScript(rawScript map[string]interface{}) persistentScript {
	script := persistentScript{}

	script.Title = rawScript["Title"].(string)
	script.Enabled = rawScript["Enabled"].(bool)
	script.Type = scriptType(rawScript["Type"].(string))
	script.Body = rawScript["Body"].(string)

	return script
}

func flattenScript(script persistentScript) map[string]interface{} {
	rawScript := make(map[string]interface{})

	rawScript["Title"] = script.Title
	rawScript["Enabled"] = script.Enabled
	rawScript["Type"] = script.Type
	rawScript["Body"] = script.Body

	return rawScript
}

func (repo *scriptRepository) FetchAll() []persistentScript {
	col := repo.database.Use("Scripts")
	scripts := make([]persistentScript, 0)
	col.ForEachDoc(func(id int, doc []byte) (moveOn bool) {
		moveOn = true

		script := persistentScript{}
		json.Unmarshal(doc, &script)
		script.ID = id

		scripts = append(scripts, script)

		return
	})

	return scripts
}

func (repo *scriptRepository) Fetch(id int) persistentScript {
	col := repo.database.Use("Scripts")

	rawScript, err := col.Read(id)
	if err != nil {
		panic(err)
	}
	script := hydrateScript(rawScript)
	script.ID = id

	return script
}

func (repo *scriptRepository) Save(script persistentScript) {
	col := repo.database.Use("Scripts")
	data := flattenScript(script)

	if script.ID <= 0 {
		id, _ := col.Insert(data)
		script.ID = id

	} else {
		col.Update(script.ID, data)
	}
}

func (repo *scriptRepository) Delete(id int) {
	col := repo.database.Use("Scripts")
	col.Delete(id)
}

type configRepository struct {
	database *db.DB
}

func flattenConfig(config *Configuration) map[string]interface{} {
	return map[string]interface{}{
		"Network":   config.Network,
		"Nick":      config.Nick,
		"Username":  config.Username,
		"Password":  config.Password,
		"OwnerNick": config.OwnerNick,
		"OwnerHost": config.OwnerHost,
	}
}

func (repo *configRepository) FetchInto(config *Configuration) {
	col := repo.database.Use("Settings")
	col.ForEachDoc(func(id int, doc []byte) (moveOn bool) {
		moveOn = false

		json.Unmarshal(doc, config)
		config.ID = id

		return
	})
}

func (repo *configRepository) Save(config *Configuration) {
	col := repo.database.Use("Settings")
	data := flattenConfig(config)

	if config.ID <= 0 {
		id, _ := col.Insert(data)
		config.ID = id

	} else {
		col.Update(config.ID, data)
	}
}
