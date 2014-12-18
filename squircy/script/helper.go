package script

import (
	"fmt"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"regexp"
)

type httpHelper struct{}

func (client *httpHelper) Get(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
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

type dataHelper struct {
	d map[string]interface{}
}

func (db *dataHelper) Get(key string) interface{} {
	if val, ok := db.d[key]; ok {
		return val
	}

	return nil
}

func (db *dataHelper) Set(key string, val interface{}) {
	db.d[key] = val
}

type ircHelper struct {
	manager *irc.IrcConnectionManager
}

func (h *ircHelper) Privmsg(target, message string) {
	conn := h.manager.Connection()
	if conn == nil {
		return
	}
	conn.Privmsg(target, message)
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

type scriptHelper struct {
	e          event.EventManager
	jsDriver   javascriptDriver
	handlers   map[string]event.EventHandler
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

type osHelper struct {}

type sysInfo struct {
	CPU string
	Cores uint
	Threads uint
	User float32
	Sys float32
	Idle float32
	MemTotal uint
	MemFree uint
	SwapTotal uint
	SwapFree uint
	Uptime uint
}

func getCommandOutput(command string, args... string) string {
	out, _ := exec.Command(command, args...).Output()

	return strings.TrimRight(string(out), "\n")
}

var vmStatSplitter = regexp.MustCompile(`:\s+`)
var swapMatcher = regexp.MustCompile(`\d+`)
var topMatcher = regexp.MustCompile(`CPU usage: ([0-9\.]+)% user, ([0-9\.]+)% sys, ([0-9\.]+)% idle`)

func getDarwinSysInfo() sysInfo {
	info := sysInfo{}

	info.CPU = getCommandOutput("sysctl", "-n", "machdep.cpu.brand_string")

	cores, _ := strconv.ParseInt(getCommandOutput("sysctl", "-n", "machdep.cpu.core_count"), 10, 32)
	info.Cores = uint(cores)

	threads, _ := strconv.ParseInt(getCommandOutput("sysctl", "-n", "machdep.cpu.thread_count"), 10, 32)
	info.Threads = uint(threads)

	memTotal, _ := strconv.ParseInt(getCommandOutput("sysctl", "-n", "hw.memsize"), 10, 64)
	info.MemTotal = uint(memTotal)

	vmStats := strings.Split(getCommandOutput("vm_stat"), "\n")
	for _, stat := range vmStats {
		parts := vmStatSplitter.Split(stat, 2)
		switch {
		case parts[0] == "Pages free" || parts[0] == "Pages speculative":
			value, _ := strconv.ParseInt(strings.TrimRight(parts[1], "."), 10, 64)
			info.MemFree += (uint(value) * 4096)
		}
	}

	pageInfo := getCommandOutput("sysctl", "-n", "vm.swapusage")
	matches := swapMatcher.FindAllString(pageInfo, -1)

	if (len(matches) > 0) {
		swapTotal, _ := strconv.ParseInt(matches[0], 10, 64)
		info.SwapTotal = uint(swapTotal)*1024*1024
	}

	if (len(matches) > 4) {
		swapFree, _ := strconv.ParseInt(matches[4], 10, 64)
		info.SwapFree = uint(swapFree) * 1024 * 1024
	}

	top := getCommandOutput("top", "-l1", "-n0")
	matches = topMatcher.FindStringSubmatch(top)

	user, _ := strconv.ParseFloat(matches[1], 64)
	sys, _ := strconv.ParseFloat(matches[2], 64)
	idle, _ := strconv.ParseFloat(matches[3], 64)

	info.User = float32(user)
	info.Sys = float32(sys)
	info.Idle = float32(idle)

	uptime, _ := strconv.ParseInt(swapMatcher.FindString(getCommandOutput("sysctl", "-n", "kern.boottime")), 10, 32)
	info.Uptime = uint(uptime)

	return info
}

func (h *osHelper) SysInfo() sysInfo {
	out, _ := exec.Command("uname").Output()
	uname := strings.TrimRight(string(out), "\n")
	switch {
	case uname == "Darwin":
		return getDarwinSysInfo()
	}

	return sysInfo{}
}
