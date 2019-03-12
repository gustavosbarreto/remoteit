package main

import "github.com/coreos/go-systemd/sdjournal"

type LogWatcher struct {
	journal *sdjournal.Journal
}
