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
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"
)

const (
	VERSION          = "0.1"
	DEFAULT_REFRESH  = 5 // default refresh interval in seconds
	DEFAULT_WEB_ADDR = "0.0.0.0:8080"
	HISTORY_LENGTH   = 10 * 60 / DEFAULT_REFRESH // for 10 minutes
)

var (
	currentUser   *user.User
	sshConfigRead bool
	allStats      *HostStats
)

func usage(code int) {
	fmt.Printf(
		`rtop-vis %s - (c) 2015 RapidLoop - MIT Licensed - http://rtop-monitor.org/rtop-vis
rtop-vis monitors system stats for a cluster over SSH

Usage: rtop-vis host [host ...]

    host
        one or more host to monitor, "ssh host" should work without password

After invoking, web UI will be available on http://localhost:8080/. Stats will
be collected every 5 seconds and graphs will refresh every 10 seconds. Graphs
will show 10 minutes of history.
`, VERSION)
	os.Exit(code)
}

func main() {

	if len(os.Args) == 1 {
		usage(1)
	}

	log.SetPrefix("rtop-vis: ")
	log.SetFlags(0)

	// get current user
	var err error
	currentUser, err = user.Current()
	if err != nil {
		log.Print(err)
		return
	}

	// read from ~/.ssh/config if possible
	sshConfig := filepath.Join(currentUser.HomeDir, ".ssh", "config")
	if _, err := os.Stat(sshConfig); err == nil {
		sshConfigRead = parseSshConfig(sshConfig)
	}

	// start connecting
	allStats = NewHostStats(HISTORY_LENGTH)
	for _, host := range os.Args[1:] {
		go doHost(host)
	}

	// start the web server
	go startWeb()

	// wait for ^C
	ch := make(chan os.Signal, 1)
	signal.Notify(ch)
	for s := range ch {
		if s == syscall.SIGTERM || s == os.Interrupt {
			break
		}
	}
	signal.Stop(ch)
	close(ch)
}

func doHost(host string) {

	var (
		port          int
		username, key string
	)

	if sshConfigRead {
		shost, sport, suser, skey := getSshEntry(host)
		if len(shost) > 0 {
			host = shost
		}
		if sport != 0 {
			port = sport
		}
		if len(suser) > 0 {
			username = suser
		}
		if len(skey) > 0 {
			key = skey
		}
	}

	// fill in still-unknown ones with defaults
	if port == 0 {
		port = 22
	}
	if len(username) == 0 {
		username = currentUser.Username
	}
	if len(key) == 0 {
		idrsap := filepath.Join(currentUser.HomeDir, ".ssh", "id_rsa")
		if _, err := os.Stat(idrsap); err == nil {
			key = idrsap
		}
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client := sshConnect(username, addr, key)
	if client == nil {
		return
	}

	for {
		stats := Stats{At: time.Now(), Hostname: host}
		getAllStats(client, &stats)
		allStats.GetRing(stats.Hostname).Add(stats)
		time.Sleep(DEFAULT_REFRESH * time.Second)
	}
}
