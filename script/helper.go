package script

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/veonik/squircy2/config"
	"github.com/veonik/squircy2/event"
	"github.com/veonik/squircy2/irc"
	"github.com/veonik/squircy2/sysinfo"
)

type httpHelper struct{}

func (client *httpHelper) Get(uri string, headers ...string) string {
	h := map[string][]string{}
	for _, v := range headers {
		p := strings.Split(v, ":")
		if len(p) != 2 {
			continue
		}
		if _, ok := h[p[0]]; !ok {
			h[p[0]] = make([]string, 0)
		}
		h[p[0]] = append(h[p[0]], p[1])
	}
	req := &http.Request{
		Method: "GET",
		Header: http.Header(h),
	}
	req.URL, _ = url.Parse(uri)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(b)
}

func (client *httpHelper) Post(uri string, body string, headers ...string) string {
	h := map[string][]string{}
	for _, v := range headers {
		p := strings.Split(v, ":")
		if len(p) != 2 {
			continue
		}
		if _, ok := h[p[0]]; !ok {
			h[p[0]] = make([]string, 0)
		}
		h[p[0]] = append(h[p[0]], p[1])
	}
	req := &http.Request{
		Method:        "POST",
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		Header:        http.Header(h),
		ContentLength: int64(len(body)),
	}
	req.URL, _ = url.Parse(uri)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}

type configHelper struct {
	conf *config.Configuration
}

func (h *configHelper) OwnerNick() string {
	return h.conf.OwnerNick
}

func (h *configHelper) OwnerHost() string {
	return h.conf.OwnerHost
}

type ircHelper struct {
	manager *irc.ConnectionManager
}

func (h *ircHelper) Privmsg(target, message string) {
	conn := h.manager.Connection()
	if conn == nil {
		return
	}
	conn.Privmsg(target, message)
}

func (h *ircHelper) Action(target, message string) {
	conn := h.manager.Connection()
	if conn == nil {
		return
	}
	conn.Action(target, message)
}

func (h *ircHelper) Join(target string) {
	conn := h.manager.Connection()
	if conn == nil {
		return
	}
	conn.Join(target)
}

func (h *ircHelper) Part(target string) {
	conn := h.manager.Connection()
	if conn == nil {
		return
	}
	conn.Part(target)
}

func (h *ircHelper) CurrentNick() string {
	conn := h.manager.Connection()
	if conn == nil {
		return ""
	}
	return conn.GetNick()
}

func (h *ircHelper) Nick(newNick string) {
	conn := h.manager.Connection()
	if conn == nil {
		return
	}
	conn.Nick(newNick)
}

func (h *ircHelper) Raw(command string) {
	conn := h.manager.Connection()
	if conn == nil {
		return
	}
	conn.SendRaw(command)
}

type scriptHelper struct {
	e        event.EventManager
	jsDriver javascriptDriver
	handlers map[string]event.EventHandler
}

func handlerId(scriptType ScriptType, eventType event.EventType, fnName string) string {
	return fmt.Sprintf("%v-%v-%v", scriptType, eventType, fnName)
}

// Bind adds a handler of the given script type for the given event type
func (s *scriptHelper) Bind(scriptType ScriptType, eventType event.EventType, fnName string) {
	id := handlerId(scriptType, eventType, fnName)
	var d scriptDriver
	switch {
	case scriptType == Javascript:
		d = s.jsDriver
	}

	handler := func(ev event.Event) {
		d.Handle(ev, fnName)
	}
	s.handlers[id] = handler
	s.e.Bind(eventType, handler)
}

// Unbind removes a handler of the given script type for the given event type
func (s *scriptHelper) Unbind(scriptType ScriptType, eventType event.EventType, fnName string) {
	id := handlerId(scriptType, eventType, fnName)
	handler, ok := s.handlers[id]
	if !ok {
		return
	}
	s.e.Unbind(eventType, handler)
	delete(s.handlers, id)
}

func (s *scriptHelper) Trigger(eventType event.EventType, data map[string]interface{}) {
	s.e.Trigger(eventType, data)
}

type osHelper struct{}

func (h *osHelper) SystemInfo() sysinfo.SystemInfo {
	return sysinfo.New()
}

type mathHelper struct{}

func (h *mathHelper) Rand() float64 {
	return rand.Float64()
}

func (h *mathHelper) Round(v float64) int {
	return int(math.Floor(v + .5))
}

func (h *mathHelper) Ceil(v float64) int {
	return int(math.Ceil(v))
}

func (h *mathHelper) Floor(v float64) int {
	return int(math.Floor(v))
}

type fileHelper struct {
	conf *config.Configuration
}

func (h *fileHelper) ReadAll(name string) (string, error) {
	if !h.conf.EnableFileAPI {
		return "", errors.New("file: file api is disabled")
	}
	p := filepath.Clean(filepath.Join(h.conf.FileAPIRoot, name))
	if !strings.HasPrefix(p, h.conf.FileAPIRoot) {
		return "", fmt.Errorf("file: path does not exist within configured root: %s", p)
	}
	res, err := ioutil.ReadFile(p)
	return string(res), err
}
