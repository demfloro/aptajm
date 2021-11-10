package main

import (
	"context"
	"log"
	"log/syslog"

	"golang.org/x/text/encoding/charmap"
)

const (
	proto = "tcp"
)

func main() {
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()
	logger, err := syslog.NewLogger(syslog.LOG_DEBUG, 0)
	if err != nil {
		log.Fatal(err)
	}

	charmap := charmap.Windows1251
	bot, err := newIRCBot(rootCtx, Config.DBName, Config.Nick,
		Config.Ident, Config.Realname, Config.Password,
		Config.NickservPass, proto, Config.Server, charmap, logger)
	if err != nil {
		logger.Fatal(err)
	}
	defer bot.Quit()
	for _, channel := range Config.Channels {
		ctx, cancel := context.WithTimeout(rootCtx, Config.Timeout)
		_, err = bot.Join(ctx, channel)
		cancel()
		if err != nil {
			logger.Printf("Error joining channel %q: %q", channel, err)
		}
	}
	bot.Wait()
}
