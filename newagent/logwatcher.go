package main

import (
	"io"
	"strconv"
	"time"

	"github.com/coreos/go-systemd/journal"
	"github.com/coreos/go-systemd/sdjournal"
	"github.com/sirupsen/logrus"
)

type LogEntry struct {
	Message   string
	Timestamp time.Time
	Level     string
}

type LogWatcher struct {
	journal *sdjournal.Journal

	ch chan *LogEntry
}

func NewLogWatcher() (*LogWatcher, error) {
	l := &LogWatcher{
		ch: make(chan *LogEntry),
	}

	var err error

	l.journal, err = sdjournal.NewJournal()
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (l *LogWatcher) Watch() <-chan *LogEntry {
	go l.watchLoop()
	return l.ch
}

func (l *LogWatcher) watchLoop() {
	defer func() {
		l.journal.Close()
	}()

	for {
		n, err := l.journal.Next()
		if err != nil && err != io.EOF {
			logrus.Error(err)
			break
		}

		if n < 1 {
			l.journal.Wait(sdjournal.IndefiniteWait)
			continue
		}

		e, err := l.journal.GetEntry()
		if err != nil {
			logrus.WithFields(logrus.Fields{"err": err}).Error("Failed to get journal entry")
			continue
		}

		if e.Fields[sdjournal.SD_JOURNAL_FIELD_PRIORITY] == "" {
			e.Fields[sdjournal.SD_JOURNAL_FIELD_PRIORITY] = "6"
		}

		level, err := strconv.Atoi(e.Fields[sdjournal.SD_JOURNAL_FIELD_PRIORITY])
		if err != nil {
			logrus.WithFields(logrus.Fields{"err": err}).Error("Failed to convert journal entry priority")
			continue
		}

		levels := map[journal.Priority]string{
			journal.PriEmerg:   "emerg",
			journal.PriAlert:   "alert",
			journal.PriCrit:    "crit",
			journal.PriErr:     "err",
			journal.PriWarning: "warning",
			journal.PriNotice:  "notice",
			journal.PriInfo:    "info",
			journal.PriDebug:   "debug",
		}

		l.ch <- &LogEntry{
			Message:   e.Fields[sdjournal.SD_JOURNAL_FIELD_MESSAGE],
			Timestamp: time.Unix(0, int64(e.RealtimeTimestamp*1000)),
			Level:     levels[journal.Priority(level)],
		}
	}
}
