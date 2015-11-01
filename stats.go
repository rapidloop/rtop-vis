/*

rtop-vis - ad hoc cluster monitoring over SSH

Copyright (c) 2015 RapidLoop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"bufio"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type Stats struct {
	At         time.Time
	Hostname   string
	Load1      float64
	MemUsed    uint64
	MemTotal   uint64
	MemFree    uint64
	MemBuffers uint64
	MemCached  uint64
}

func getAllStats(host string, client *ssh.Client, stats *Stats) {
	getHostname(host, client, stats)
	getLoad(client, stats)
	getMemInfo(client, stats)
}

func getHostname(host string, client *ssh.Client, stats *Stats) (err error) {
	hostname, err := runCommand(client, "/bin/hostname -f")
	if err != nil {
		stats.Hostname = strings.TrimSpace(host)
		return
	}
	stats.Hostname = strings.TrimSpace(hostname)
	return
}

func getLoad(client *ssh.Client, stats *Stats) (err error) {
	line, err := runCommand(client, "/bin/cat /proc/loadavg")
	if err != nil {
		return
	}

	parts := strings.Fields(line)
	if len(parts) == 5 {
		if f, err := strconv.ParseFloat(parts[0], 64); err == nil {
			stats.Load1 = f
		}
	}

	return
}

func getMemInfo(client *ssh.Client, stats *Stats) (err error) {
	lines, err := runCommand(client, "/bin/cat /proc/meminfo")
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(lines))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 3 {
			val, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				continue
			}
			val *= 1024
			switch parts[0] {
			case "MemTotal:":
				stats.MemTotal = val
			case "MemFree:":
				stats.MemFree = val
			case "Buffers:":
				stats.MemBuffers = val
			case "Cached:":
				stats.MemCached = val
			}
		}
	}

	stats.MemUsed = stats.MemTotal - (stats.MemFree + stats.MemBuffers + stats.MemCached)
	return
}

//--
// A ring buffer of Stats objects. Kind of.

type StatsRing struct {
	sync.Mutex
	Values []Stats
	Head   int
}

func NewStatsRing(n int) *StatsRing {
	return &StatsRing{Values: make([]Stats, n)}
}

func (r *StatsRing) Add(s Stats) {
	r.Lock()
	defer r.Unlock()
	r.Values[r.Head] = s
	r.Head = (r.Head + 1) % len(r.Values)
}

// oldest first
func (r *StatsRing) Entries() []Stats {
	r.Lock()
	defer r.Unlock()
	s := make([]Stats, 0, len(r.Values))
	pos := r.Head
	for {
		pos = (pos + 1) % len(r.Values)
		if pos == r.Head {
			return s
		}
		if !r.Values[pos].At.IsZero() {
			s = append(s, r.Values[pos])
		}
	}
}

//--
// A ring buffer for each host.

type HostStats struct {
	sync.Mutex
	Map   map[string]*StatsRing
	Count int
}

func NewHostStats(count int) *HostStats {
	return &HostStats{
		Map:   make(map[string]*StatsRing),
		Count: count,
	}
}

func (h *HostStats) GetRing(host string) *StatsRing {
	h.Lock()
	defer h.Unlock()
	if r, found := h.Map[host]; found {
		return r
	} else {
		r = NewStatsRing(h.Count)
		h.Map[host] = r
		return r
	}
}

func (h *HostStats) Keys() []string {
	h.Lock()
	defer h.Unlock()
	keys := make([]string, 0, len(h.Map))
	for k, _ := range h.Map {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
