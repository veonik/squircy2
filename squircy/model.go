package squircy

import (
	"encoding/json"
	"fmt"
	"github.com/fzzy/radix/redis"
)

const deleteRecord = "<<delete>>"

type scriptType string

const (
	scriptJavascript scriptType = "Javascript"
	scriptLua                   = "Lua"
	scriptLisp                  = "Lisp"
)

type persistentScript struct {
	Type    scriptType
	Title   string
	Body    string
	Enabled bool
}

type scriptRepository struct {
	client *redis.Client
}

func hydrateScript(rawScript string) (persistentScript, error) {
	script := persistentScript{}
	err := json.Unmarshal(([]byte(rawScript)), &script)
	if err != nil {
		fmt.Println(err)
		return script, err
	}

	return script, nil
}

func flattenScript(script persistentScript) (string, error) {
	str, err := json.Marshal(&script)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}

	return string(str[0:len(str)]), nil
}

func (repo *scriptRepository) Fetch() []persistentScript {
	data, err := repo.client.Cmd("lrange", "scripts", 0, -1).List()
	if err != nil {
		panic(err)
	}
	fmt.Println(data)

	scripts := make([]persistentScript, 0)
	for _, rawScript := range data {
		fmt.Println(rawScript)
		script, err := hydrateScript(rawScript)
		if err != nil {
			continue
		}

		scripts = append(scripts, script)
	}

	return scripts
}

func (repo *scriptRepository) FetchIndex(index int) persistentScript {
	rawScript := repo.client.Cmd("lindex", "scripts", index).String()
	script, _ := hydrateScript(rawScript)

	return script
}

func (repo *scriptRepository) Save(index int, script persistentScript) {
	data, _ := flattenScript(script)
	if index < 0 {
		repo.client.Cmd("rpush", "scripts", data)
	} else {
		repo.client.Cmd("lset", "scripts", index, data)
	}
}

func (repo *scriptRepository) Delete(index int) {
	repo.client.Cmd("lset", "scripts", index, deleteRecord)
	repo.client.Cmd("lrem", "scripts", 0, deleteRecord)
}
