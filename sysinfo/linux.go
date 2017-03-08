// +build linux

package sysinfo

import (
	"regexp"
	"strconv"
)

var modelMatcher = regexp.MustCompile(`Model name:\s+(.+)`)
var coresMatcher = regexp.MustCompile(`Core\(s\) per socket:\s+(.+)`)
var socketsMatcher = regexp.MustCompile(`Socket\(s\):\s+(.+)`)
var threadsMatcher = regexp.MustCompile(`Thread\(s\) per core:\s+(.+)`)
var topMatcher = regexp.MustCompile(`([0-9.]+)%?\s?us,\s+([0-9.]+)%?\s?sy,.+?([0-9.]+)%?\s?id`)
var memMatcher = regexp.MustCompile(`Mem:\s+([0-9]+)\s+[0-9]+\s+([0-9]+)`)
var swapMatcher = regexp.MustCompile(`Swap:\s+([0-9]+)\s+[0-9]+\s+([0-9]+)`)
var uptimeMatcher = regexp.MustCompile(`up\s([0-9]+).+?\s([0-9]+):([0-9]+)`)

func New() (s SystemInfo) {
	s = SystemInfo{}
	c := newExecutor()

	cpu := c.getOutput("lscpu")
	matches := modelMatcher.FindStringSubmatch(cpu)
	if len(matches) == 2 {
		s.CPU = matches[1]
	}

	var cores uint64 = 0
	matches = coresMatcher.FindStringSubmatch(cpu)
	if len(matches) == 2 {
		cores, _ = strconv.ParseUint(matches[1], 10, 64)
	}

	var sockets uint64 = 0
	matches = socketsMatcher.FindStringSubmatch(cpu)
	if len(matches) == 2 {
		sockets, _ = strconv.ParseUint(matches[1], 10, 64)
	}

	var threads uint64 = 0
	matches = threadsMatcher.FindStringSubmatch(cpu)
	if len(matches) == 2 {
		threads, _ = strconv.ParseUint(matches[1], 10, 64)
	}

	s.Cores = uint(cores * sockets)
	s.Threads = uint(cores * threads * sockets)

	top := c.getOutput("top", "-bn1")
	matches = topMatcher.FindStringSubmatch(top)
	if len(matches) == 4 {
		user, _ := strconv.ParseFloat(matches[1], 64)
		sys, _ := strconv.ParseFloat(matches[2], 64)
		idle, _ := strconv.ParseFloat(matches[3], 64)

		s.User = float32(user)
		s.Sys = float32(sys)
		s.Idle = float32(idle)
	}

	free := c.getOutput("free", "-m")
	matches = memMatcher.FindStringSubmatch(free)
	if len(matches) == 3 {
		tot, _ := strconv.ParseUint(matches[1], 10, 64)
		fre, _ := strconv.ParseUint(matches[2], 10, 64)

		s.MemTotal = uint(tot * 1024)
		s.MemFree = uint(fre * 1024)
	}

	matches = swapMatcher.FindStringSubmatch(free)
	if len(matches) == 3 {
		tot, _ := strconv.ParseUint(matches[1], 10, 64)
		fre, _ := strconv.ParseUint(matches[2], 10, 64)

		s.SwapTotal = uint(tot * 1024)
		s.SwapFree = uint(fre * 1024)
	}

	uptime := c.getOutput("uptime")
	matches = uptimeMatcher.FindStringSubmatch(uptime)
	if len(matches) == 4 {
		days, _ := strconv.ParseUint(matches[1], 10, 64)
		hours, _ := strconv.ParseUint(matches[2], 10, 64)
		mins, _ := strconv.ParseUint(matches[3], 10, 64)
		s.Uptime = uint((days * 86400) + (hours * 3600) + (mins * 60))
	}
	return
}
