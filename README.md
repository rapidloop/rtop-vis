
# rtop-vis

[![Join the chat at https://gitter.im/rapidloop/rtop-vis](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/rapidloop/rtop-vis?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

`rtop-vis` can monitor load and memory usage of all the specified servers
and visualize the data as a graph with a bit (10 minutes' worth) of history.
It connects to servers via SSH and does not need anything to be installed
on the servers. The collected data is not persisted. It is lost when `rtop-vis` is stopped.

`rtop-vis` is MIT-licensed and can be used anywhere with attribution.

*`rtop-vis`'s [home page](http://www.rtop-monitor.org/rtop-vis) has more
information and screenshots!*

You might also be interested in the sibling projects
[rtop](http://www.rtop-monitor.org/) (CLI version, no GUI, single server)
and [rtop-bot](http://www.rtop-monitor.org/rtop-bot) (Slack and HipChat
bot version, on-demand).

## build

`rtop-vis` is written in [go](http://golang.org/), and requires Go
version 1.2 or higher. To build, `go get` it:

    go get github.com/rapidloop/rtop-vis

You should find the binary `rtop-vis` under `$GOPATH/bin` when the command
completes. There are no runtime dependencies or configuration needed.

## contribute

Pull requests welcome. Keep it simple.

## changelog
* 13-Sep-2015: first public release
