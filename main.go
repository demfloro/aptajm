package main

import (
	"context"
	"log"
	"os"
	"time"
)

func main() {
	var err error
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()
	logfile, err := os.OpenFile("/var/log/gobot/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Print(err)
		return
	}
	logger := newLogger(logfile, true)
	if len(os.Args) < 2 {
		log.Printf("Usage: %s configfile", os.Args[0])
		return
	}
	config, err := loadConfig(os.Args[1])
	if err != nil {
		logger.Logf("Error during loading config: %q", err)
		return
	}
	for {
		botCtx, botCancel := context.WithCancel(rootCtx)
		bot, err := newIRCBot(botCtx, *config, logger)
		if err != nil {
			logger.Log(err)
			return
		}
		err = bot.Wait()
		if err == nil {
			botCancel()
			break
		}
		// attempt to reconnect
		bot.Quit()
		botCancel()
		time.Sleep(10 * time.Second)
	}
}
