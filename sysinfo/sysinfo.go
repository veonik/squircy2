package sysinfo // import "github.com/veonik/squircy2/sysinfo"

import (
	"os/exec"
	"strconv"
	"strings"
)

type SystemInfo struct {
	CPU       string
	Cores     uint
	Threads   uint
	User      float32
	Sys       float32
	Idle      float32
	MemTotal  uint
	MemFree   uint
	SwapTotal uint
	SwapFree  uint
	Uptime    uint
}

type executor struct{}

func newExecutor() executor {
	return executor{}
}

func (c *executor) getOutput(command string, args ...string) string {
	out, err := exec.Command(command, args...).Output()
	if err != nil {
		return ""
	}

	return string(out)
}

func (c *executor) getString(command string, args ...string) string {
	out, err := exec.Command(command, args...).Output()
	if err != nil {
		return ""
	}

	return strings.TrimRight(string(out), "\n")
}

func (c *executor) getStrings(command string, args ...string) []string {
	out, err := exec.Command(command, args...).Output()
	if err != nil {
		return make([]string, 0)
	}

	return strings.Split(string(out), "\n")
}

func (c *executor) getUInt(command string, args ...string) uint {
	out, err := strconv.ParseInt(c.getString(command, args...), 10, 64)
	if err != nil {
		return 0
	}

	return uint(out)
}

func (c *executor) getFloat32(command string, args ...string) float32 {
	out, err := strconv.ParseFloat(c.getString(command, args...), 32)
	if err != nil {
		return 0
	}

	return float32(out)
}
