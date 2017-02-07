package script

import (
	"encoding/json"
	"sort"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"io/ioutil"
	"fmt"
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
	conf *config.Configuration
	files *fileRepository
	db *dbRepository
}

type fileRepository struct {
	conf *config.Configuration
}

type dbRepository struct {
	database *db.DB
}

func NewScriptRepository(database *db.DB, conf *config.Configuration) *ScriptRepository {
	return &ScriptRepository{conf, &fileRepository{conf}, &dbRepository{database}}
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
	if repo.conf.ScriptsAsFiles {
		return repo.files.FetchAll()
	}
	return repo.db.FetchAll()
}

func (repo *ScriptRepository) Fetch(id int) *Script {
	if repo.conf.ScriptsAsFiles {
		return repo.files.Fetch(id)
	}
	return repo.db.Fetch(id)
}

func (repo *ScriptRepository) Save(script *Script) {
	if repo.conf.ScriptsAsFiles {
		repo.files.Save(script)
		return
	}
	repo.db.Save(script)
}

func (repo *ScriptRepository) Delete(id int) {
	if repo.conf.ScriptsAsFiles {
		repo.files.Delete(id)
	}
	repo.db.Delete(id)
}


func (repo *dbRepository) FetchAll() []*Script {
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

func (repo *dbRepository) Fetch(id int) *Script {
	col := repo.database.Use("Scripts")

	rawScript, err := col.Read(id)
	if err != nil {
		panic(err)
	}
	script := hydrateScript(rawScript)
	script.ID = id

	return script
}

func (repo *dbRepository) Save(script *Script) {
	col := repo.database.Use("Scripts")
	data := flattenScript(script)

	if script.ID <= 0 {
		id, _ := col.Insert(data)
		script.ID = id

	} else {
		col.Update(script.ID, data)
	}
}

func (repo *dbRepository) Delete(id int) {
	col := repo.database.Use("Scripts")
	col.Delete(id)
}


func (repo *fileRepository) FetchAll() []*Script {
	var scripts []*Script
	files, err := ioutil.ReadDir(repo.conf.ScriptsPath)
	if err != nil {
		return scripts
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		contents, err := ioutil.ReadFile(file.Name())
		if err != nil {
			fmt.Println(file)
			scripts = append(scripts, &Script{
				Title: file.Name(),
				Body: string(contents),
				Type: Javascript,
				ID: -1,
			})
		}
	}
	return scripts
}

func (repo *fileRepository) Fetch(id int) *Script {


	return &Script{Type: Javascript}
}

func (repo *fileRepository) Save(script *Script) {

}

func (repo *fileRepository) Delete(id int) {

}