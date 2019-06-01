package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
)

type DB struct {
	rootPath string

	logger *log.Logger

	open map[string]*Collection

	mu sync.Mutex
}

func OpenDB(rootPath string, l *log.Logger) (*DB, error) {
	return &DB{rootPath: rootPath, open: make(map[string]*Collection), logger: l}, nil
}

func (d *DB) Use(name string) *Collection {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.open[name]; !ok {
		dir := filepath.Join(d.rootPath, name)
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				d.logger.Warnln("failed to create directory for collection:", err)
				return nil
			}
		}
		d.open[name] = &Collection{basePath: dir, cache: make(map[int]document), logger: d.logger}
	}
	return d.open[name]
}

func (d *DB) Create(name string) error {
	return errors.New("unsupported operation")
}

type document map[string]interface{}

type Collection struct {
	basePath string
	cache    map[int]document

	logger *log.Logger

	mu sync.Mutex
}

func (c *Collection) Read(id int) (doc map[string]interface{}, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if d, ok := c.cache[id]; ok {
		return d, nil
	}
	p := filepath.Join(c.basePath, fmt.Sprintf("%d.json", id))
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	v := make(map[string]interface{})
	err = json.Unmarshal(b, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (c *Collection) Insert(doc map[string]interface{}) (int, error) {
	newId := func() int {
		for {
			id := rand.Int()
			if _, ok := c.cache[id]; ok {
				continue
			}
			_, err := os.Stat(filepath.Join(c.basePath, fmt.Sprintf("%d.json", id)))
			if !os.IsNotExist(err) {
				continue
			}
			return id
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	id := newId()
	b, err := json.Marshal(doc)
	if err != nil {
		return 0, err
	}
	p := filepath.Join(c.basePath, fmt.Sprintf("%d.json", id))
	err = ioutil.WriteFile(p, b, 0644)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (c *Collection) Update(id int, doc map[string]interface{}) error {
	b, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	p := filepath.Join(c.basePath, fmt.Sprintf("%d.json", id))
	err = ioutil.WriteFile(p, b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func EvalQuery(q interface{}, src *Collection, result *map[int]struct{}) (err error) {
	return errors.New("unsupported operation")
}

func (c *Collection) ForEachDoc(fn func(id int, doc []byte) (moveOn bool)) {
	fs, err := ioutil.ReadDir(c.basePath)
	if err != nil {
		c.logger.Warnln("failed to read data directory:", err)
		return
	}
	for _, f := range fs {
		id := new(int)
		if _, err := fmt.Sscanf(f.Name(), "%d.json", id); err != nil {
			continue
		}
		b, err := ioutil.ReadFile(filepath.Join(c.basePath, f.Name()))
		if err != nil {
			c.logger.Warnln("failed to read data file:", err)
			continue
		}
		if !fn(*id, b) {
			return
		}
	}
}

func (c *Collection) Index(cols []string) error {
	return errors.New("unsupported operation")
}

func (c *Collection) Delete(id int) {
	fp := filepath.Join(c.basePath, fmt.Sprintf("%d.json", id))
	_, err := os.Stat(fp)
	if os.IsNotExist(err) {
		return
	}
	c.logger.Errorln(os.Remove(fp))
}
