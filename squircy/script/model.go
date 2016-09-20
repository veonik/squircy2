package script

import (
	"encoding/json"
	"sort"

	"github.com/HouzuoGuo/tiedot/db"
)

type ScriptType string

const (
	Javascript ScriptType = "Javascript"
)

type Script struct {
	ID      int
	Type    ScriptType
	Title   string
	Body    string
	Enabled bool
}

type ScriptRepository struct {
	database *db.DB
}

func NewScriptRepository(database *db.DB) ScriptRepository {
	return ScriptRepository{database}
}

func hydrateScript(rawScript map[string]interface{}) *Script {
	script := &Script{}

	script.Title = rawScript["Title"].(string)
	script.Enabled = rawScript["Enabled"].(bool)
	script.Type = ScriptType(rawScript["Type"].(string))
	script.Body = rawScript["Body"].(string)
	script.Enabled = rawScript["Enabled"].(bool)

	return script
}

func flattenScript(script *Script) map[string]interface{} {
	rawScript := make(map[string]interface{})

	rawScript["Title"] = script.Title
	rawScript["Enabled"] = script.Enabled
	rawScript["Type"] = script.Type
	rawScript["Body"] = script.Body
	rawScript["Enabled"] = script.Enabled

	return rawScript
}

type scriptSlice []*Script

func (s scriptSlice) Len() int {
	return len(s)
}

func (s scriptSlice) Less(i, j int) bool {
	return s[i].Title < s[j].Title
}

func (s scriptSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (repo *ScriptRepository) FetchAll() []*Script {
	col := repo.database.Use("Scripts")
	scripts := make([]*Script, 0)
	col.ForEachDoc(func(id int, doc []byte) (moveOn bool) {
		moveOn = true

		val := make(map[string]interface{}, 0)

		json.Unmarshal(doc, &val)
		script := hydrateScript(val)
		script.ID = id

		scripts = append(scripts, script)

		return
	})

	sort.Sort(scriptSlice(scripts))

	return scripts
}

func (repo *ScriptRepository) Fetch(id int) *Script {
	col := repo.database.Use("Scripts")

	rawScript, err := col.Read(id)
	if err != nil {
		panic(err)
	}
	script := hydrateScript(rawScript)
	script.ID = id

	return script
}

func (repo *ScriptRepository) Save(script *Script) {
	col := repo.database.Use("Scripts")
	data := flattenScript(script)

	if script.ID <= 0 {
		id, _ := col.Insert(data)
		script.ID = id

	} else {
		col.Update(script.ID, data)
	}
}

func (repo *ScriptRepository) Delete(id int) {
	col := repo.database.Use("Scripts")
	col.Delete(id)
}
