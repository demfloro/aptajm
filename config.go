package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var Config ConfigStruct

type ConfigStruct struct {
	Nick, Password, Ident     string
	Realname, WeatherToken    string
	Server, DBName            string
	UserAgent, NickservPass   string
	Admins, Channels, Ignored []string
	BackupTimeout, Timeout    time.Duration
}

func init() {
	if len(os.Args) < 2 {
		log.Fatalf("Not enough arguments, usage: %s configfile", os.Args[0])
	}
	Config = loadConfig(os.Args[1])
}

func loadConfig(fname string) (c ConfigStruct) {
	file, err := os.Open(fname)
	defer file.Close()
	if err != nil {
		log.Fatalf("Failed to open configuration file: %q", err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		splitted := strings.Split(line, "=")
		if len(splitted) != 2 {
			log.Fatalf("Failed to parse: %q", line)
		}
		k, v := strings.ToLower(splitted[0]), splitted[1]
		if strings.HasPrefix(k, "#") {
			continue
		}
		switch k {
		case "nick":
			if c.Nick != "" {
				log.Fatalf("Repeated Nick assignment")
			}
			c.Nick = v
		case "password":
			if c.Password != "" {
				log.Fatalf("Repeated Password assignment")
			}
			c.Password = v
		case "ident":
			if c.Ident != "" {
				log.Fatalf("Repeated Ident assignment")
			}
			c.Ident = v
		case "realname":
			if c.Realname != "" {
				log.Fatalf("Repeated Realname assignment")
			}
			c.Realname = v
		case "weathertoken":
			if c.WeatherToken != "" {
				log.Fatalf("Repeated Weathertoken assignment")
			}
			c.WeatherToken = v
		case "server":
			if c.Server != "" {
				log.Fatalf("Repeated Server assignment")
			}
			c.Server = v
		case "channels":
			if len(c.Channels) != 0 {
				log.Fatalf("Repeated Channels assignment")
			}
			c.Channels = strings.Split(v, ",")
		case "timeout":
			n, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				log.Fatalf("%q is not valid unsigned integer for Timeout", v)
			}
			c.Timeout = time.Duration(n) * time.Second
		case "backuptimeout":
			n, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				log.Fatalf("%q is not valid unsigned integer for BackupTimeout", v)
			}
			c.BackupTimeout = time.Duration(n) * time.Second
		case "dbname":
			if c.DBName != "" {
				log.Fatalf("Repeated DBName assignment")
			}
			c.DBName = v
		case "useragent":
			if c.UserAgent != "" {
				log.Fatalf("Repeated UserAgent assignment")
			}
			c.UserAgent = v
		case "ignored":
			if len(c.Ignored) != 0 {
				log.Fatalf("Repeated Ignored assignment")
			}
			c.Ignored = strings.Split(v, ",")
		case "admins":
			if len(c.Admins) != 0 {
				log.Fatalf("Repeated Admins assignment")
			}
			c.Admins = strings.Split(v, ",")
		case "nickservpass":
			if c.NickservPass != "" {
				log.Fatalf("Repeated Nickservpass assignment")
			}
			c.NickservPass = v
		default:
			log.Fatalf("Unknown directive %q", k)
		}
	}
	switch {
	case c.DBName == "":
		log.Fatalf("DBName must be specified")
	case c.Nick == "":
		log.Fatalf("Nick must be specified")
	case c.Ident == "":
		log.Fatalf("Ident must be specified")
	case c.Realname == "":
		log.Fatalf("Realname must be specified")
	case c.Server == "":
		log.Fatalf("Server must be specified")
	case len(c.Channels) == 0:
		log.Fatalf("Channel must be specified")
	case c.Timeout == 0:
		c.Timeout = 10 * time.Second
	}
	if scanner.Err() != nil {
		log.Fatalf("Scanner failure: %q", scanner.Err())
	}
	return c

}
