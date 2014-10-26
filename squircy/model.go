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
