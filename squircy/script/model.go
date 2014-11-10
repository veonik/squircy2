package script

import (
	"encoding/json"
	"github.com/HouzuoGuo/tiedot/db"
)

type ScriptType string

const (
	Javascript ScriptType = "Javascript"
	Lua                   = "Lua"
	Lisp                  = "Lisp"
	Anko                  = "Anko"
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

func hydrateScript(rawScript map[string]interface{}) Script {
	script := Script{}

	script.Title = rawScript["Title"].(string)
	script.Enabled = rawScript["Enabled"].(bool)
	script.Type = ScriptType(rawScript["Type"].(string))
	script.Body = rawScript["Body"].(string)

	return script
}

func flattenScript(script Script) map[string]interface{} {
	rawScript := make(map[string]interface{})

	rawScript["Title"] = script.Title
	rawScript["Enabled"] = script.Enabled
	rawScript["Type"] = script.Type
	rawScript["Body"] = script.Body

	return rawScript
}

func (repo *ScriptRepository) FetchAll() []Script {
	col := repo.database.Use("Scripts")
	scripts := make([]Script, 0)
	col.ForEachDoc(func(id int, doc []byte) (moveOn bool) {
		moveOn = true

		val := make(map[string]interface{}, 0)

		json.Unmarshal(doc, &val)
		script := hydrateScript(val)
		script.ID = id

		scripts = append(scripts, script)

		return
	})

	return scripts
}

func (repo *ScriptRepository) Fetch(id int) Script {
	col := repo.database.Use("Scripts")

	rawScript, err := col.Read(id)
	if err != nil {
		panic(err)
	}
	script := hydrateScript(rawScript)
	script.ID = id

	return script
}

func (repo *ScriptRepository) Save(script Script) {
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

type GenericRepository struct {
	database *db.DB
	coll string
}

type GenericModel map[string]interface{}

func NewGenericRepository(database *db.DB, coll string) GenericRepository {
	col := database.Use(coll)
	if col == nil {
		err := database.Create(coll)
		if err != nil {
			panic(err)
		}

		col = database.Use(coll)
	}

	return GenericRepository{database, coll}
}

func hydrateGeneric(rawGeneric map[string]interface{}) GenericModel {
	return GenericModel(rawGeneric)
}

func flattenGeneric(generic GenericModel) map[string]interface{} {
	return map[string]interface{}(generic)
}

func (repo *GenericRepository) FetchAll() []GenericModel {
	col := repo.database.Use(repo.coll)
	generics := make([]GenericModel, 0)
	col.ForEachDoc(func(id int, doc []byte) (moveOn bool) {
		moveOn = true

		val := make(map[string]interface{}, 0)

		json.Unmarshal(doc, &val)
		generic := hydrateGeneric(val)
		generic["ID"] = id

		generics = append(generics, generic)

		return
	})

	return generics
}

func (repo *GenericRepository) Fetch(id int) GenericModel {
	col := repo.database.Use(repo.coll)

	rawGeneric, err := col.Read(id)
	if err != nil {
		panic(err)
	}
	generic := hydrateGeneric(rawGeneric)
	generic["ID"] = id

	return generic
}

func (repo *GenericRepository) Save(generic GenericModel) {
	col := repo.database.Use(repo.coll)
	data := flattenGeneric(generic)

	if _, ok := generic["ID"]; !ok {
		id, _ := col.Insert(data)
		generic["ID"] = id

	} else {
		col.Update(generic["ID"].(int), data)
	}
}

func (repo *GenericRepository) Query(query interface{}) []GenericModel {
	col := repo.database.Use(repo.coll)
	result := make(map[int]struct{})
	if err := db.EvalQuery(query, col, &result); err != nil {
		panic(err)
	}

	generics := make([]GenericModel, 0)
	for id := range result {
		generics = append(generics, repo.Fetch(id))
	}

	return generics
}

func (repo *GenericRepository) Index(cols []string) {
	col := repo.database.Use(repo.coll)
	if err := col.Index(cols); err != nil {
		panic(err)
	}
}

func (repo *GenericRepository) Delete(id int) {
	col := repo.database.Use(repo.coll)
	col.Delete(id)
}
