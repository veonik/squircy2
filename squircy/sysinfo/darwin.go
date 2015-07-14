// +build darwin

package sysinfo

import (
	"regexp"
	"strconv"
	"strings"
)

var vmStatSplitter = regexp.MustCompile(`:\s+`)
var swapMatcher = regexp.MustCompile(`\d+`)
var topMatcher = regexp.MustCompile(`CPU usage: ([0-9\.]+)% user, ([0-9\.]+)% sys, ([0-9\.]+)% idle`)

func New() (s SystemInfo) {
	s = SystemInfo{}
	c := newExecutor()

	s.CPU = c.getString("sysctl", "-n", "machdep.cpu.brand_string")
	s.Cores = c.getUInt("sysctl", "-n", "machdep.cpu.core_count")
	s.Threads = c.getUInt("sysctl", "-n", "machdep.cpu.thread_count")
	s.MemTotal = c.getUInt("sysctl", "-n", "hw.memsize")

	vmStats := c.getStrings("vm_stat")
	for _, stat := range vmStats {
		parts := vmStatSplitter.Split(stat, 2)
		switch {
		case parts[0] == "Pages free" || parts[0] == "Pages speculative":
			value, _ := strconv.ParseInt(strings.TrimRight(parts[1], "."), 10, 64)
			s.MemFree += (uint(value) * 4096)
		}
	}

	pageInfo := c.getOutput("sysctl", "-n", "vm.swapusage")
	matches := swapMatcher.FindAllString(pageInfo, -1)

	if (len(matches) > 0) {
		swapTotal, _ := strconv.ParseInt(matches[0], 10, 64)
		s.SwapTotal = uint(swapTotal)*1024*1024
	}

	if (len(matches) > 4) {
		swapFree, _ := strconv.ParseInt(matches[4], 10, 64)
		s.SwapFree = uint(swapFree) * 1024 * 1024
	}

	top := c.getOutput("top", "-l1", "-n0")
	matches = topMatcher.FindStringSubmatch(top)

	user, _ := strconv.ParseFloat(matches[1], 64)
	sys, _ := strconv.ParseFloat(matches[2], 64)
	idle, _ := strconv.ParseFloat(matches[3], 64)

	s.User = float32(user)
	s.Sys = float32(sys)
	s.Idle = float32(idle)

	uptime, _ := strconv.ParseInt(swapMatcher.FindString(c.getOutput("sysctl", "-n", "kern.boottime")), 10, 32)
	s.Uptime = uint(uptime)

	return
}
