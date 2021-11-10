package main

import (
	"fmt"
	"time"
)

type quote struct {
	Id     int
	Date   time.Time
	Rating int
	Text   []string
}

func (q quote) ircFormat() (result []string) {
	id := fmt.Sprintf("\x0304#%d\x03", q.Id)
	// https://pkg.go.dev/time#pkg-constants
	date := fmt.Sprintf("\x0310%s\x03", q.Date.Format("2006-01-02 15:04"))
	rating := fmt.Sprintf("\x0304%d\x03", q.Rating)
	header := fmt.Sprintf("%s :: %s :: Rating: %s", id, date, rating)
	result = append(result, header)
	result = append(result, colorize(q.Text, "\x0303")...)
	return
}

func colorize(in []string, color string) (out []string) {
	for _, line := range in {
		out = append(out, color+line)
	}
	return out
}
