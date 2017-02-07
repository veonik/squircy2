package script

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"path/filepath"
	"sort"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/tyler-sommer/squircy2/config"
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

type ScriptRepository interface {
	FetchAll() []*Script
	Fetch(id int) *Script
	Save(script *Script)
	Delete(id int)
}

func newDBRepository(database *db.DB, logger *log.Logger) *dbRepository {
	return &dbRepository{database, logger}
}

type dbRepository struct {
	database *db.DB
	logger   *log.Logger
}

func NewScriptRepository(database *db.DB, conf *config.Configuration, logger *log.Logger) ScriptRepository {
	if conf.ScriptsAsFiles {
		return newFileRepository(conf, logger)
	}
	return newDBRepository(database, logger)
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
		// TODO: Handle error properly
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

type fileRepository struct {
	idx    map[int]string
	conf   *config.Configuration
	logger *log.Logger
}

func newFileRepository(conf *config.Configuration, logger *log.Logger) *fileRepository {
	repo := &fileRepository{make(map[int]string), conf, logger}
	repo.loadIndex()
	return repo
}

func (repo *fileRepository) loadIndex() {
	j, err := ioutil.ReadFile(filepath.Join(repo.conf.ScriptsPath, "index.json"))
	if err != nil {
		// TODO: Handle error properly
		repo.logger.Println(err.Error())
		return
	}
	err = json.Unmarshal(j, &repo.idx)
	if err != nil {
		// TODO: Handle error properly
		repo.logger.Println(err.Error())
		return
	}
}

func (repo *fileRepository) saveIndex() {
	d, err := json.Marshal(repo.idx)
	if err != nil {
		// TODO: Handle error properly
		repo.logger.Println(err.Error())
		return
	}
	err = ioutil.WriteFile(filepath.Join(repo.conf.ScriptsPath, "index.json"), d, 0644)
	if err != nil {
		// TODO: Handle error properly
		repo.logger.Println(err.Error())
		return
	}
}

func (repo *fileRepository) FetchAll() []*Script {
	var scripts []*Script
	for id, file := range repo.idx {
		contents, err := ioutil.ReadFile(filepath.Join(repo.conf.ScriptsPath, file))
		if err != nil {
			// TODO: Handle error properly
			repo.logger.Println(err.Error())
			continue
		}
		scripts = append(scripts, &Script{
			Title:   file,
			Body:    string(contents),
			Type:    Javascript,
			ID:      id,
			Enabled: true,
		})
	}
	sort.Sort(scriptSlice(scripts))
	return scripts
}

func (repo *fileRepository) Fetch(id int) *Script {
	if file, ok := repo.idx[id]; ok {
		contents, err := ioutil.ReadFile(filepath.Join(repo.conf.ScriptsPath, file))
		if err != nil {
			// TODO: Handle error properly
			repo.logger.Println(err.Error())
			return nil
		}
		return &Script{
			Title:   file,
			Body:    string(contents),
			Type:    Javascript,
			ID:      id,
			Enabled: true,
		}
	}

	return nil
}

func (repo *fileRepository) Save(script *Script) {
	if script.ID <= 0 {
		script.ID = rand.Int()
	}
	repo.idx[script.ID] = script.Title
	err := ioutil.WriteFile(filepath.Join(repo.conf.ScriptsPath, script.Title), []byte(script.Body), 0644)
	if err != nil {
		// TODO: Handle error properly
		repo.logger.Println(err.Error())
		return
	}
	repo.saveIndex()
}

func (repo *fileRepository) Delete(id int) {
	delete(repo.idx, id)
	repo.saveIndex()
}
