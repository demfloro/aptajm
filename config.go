package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	handlers = map[string]optHandler{
		"nick":           nick,
		"password":       password,
		"ident":          ident,
		"realname":       realName,
		"weathertoken":   weatherToken,
		"server":         server,
		"channels":       channels,
		"timeout":        timeout,
		"dbname":         dbname,
		"useragent":      userAgent,
		"ignored":        ignored,
		"admins":         admins,
		"nickservpass":   nickservPass,
		"pubfingerprint": pubFingerprint,
	}
)

type optHandler func(string) option
type option func(*config)

type config struct {
	nick, password, ident     string
	realname, weatherToken    string
	server, dbname            string
	userAgent, nickservPass   string
	pubFingerprint            string
	admins, channels, ignored []string
	timeout                   time.Duration
}

func loadConfig(fname string) (*config, error) {
	if len(os.Args) < 2 {
		log.Fatalf("Not enough arguments, usage: %s configfile", os.Args[0])
	}
	file, err := os.Open(fname)
	if err != nil {
		log.Fatalf("Failed to open configuration file: %q", err)
	}
	defer file.Close()
	conf, err := parseConfig(file)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func parseConfig(reader io.Reader) (c *config, err error) {
	scanner := bufio.NewScanner(reader)
	c = new(config)
	for scanner.Scan() {
		line := scanner.Text()
		splitted := strings.Split(line, "=")
		if len(splitted) != 2 {
			log.Fatalf("Failed to parse: %q", line)
		}
		k, v := strings.ToLower(strings.TrimSpace(splitted[0])), strings.TrimSpace(splitted[1])
		if strings.HasPrefix(k, "#") {
			continue
		}
		handler, ok := handlers[k]
		if !ok {
			log.Fatalf("Unknown directive %q", k)
		}
		handler(v)(c)
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	switch {
	case c.dbname == "":
		log.Fatalf("DBName must be specified")
	case c.nick == "":
		log.Fatalf("Nick must be specified")
	case c.ident == "":
		log.Fatalf("Ident must be specified")
	case c.realname == "":
		log.Fatalf("Realname must be specified")
	case c.server == "":
		log.Fatalf("Server must be specified")
	case len(c.channels) == 0:
		log.Fatalf("At least one channel in Channels must be specified")
	case c.timeout == time.Duration(0):
		c.timeout = 10 * time.Second
	}
	return c, nil

}

func nick(value string) option {
	return func(c *config) {
		if c.nick != "" {
			log.Fatalf("Repeated Nick assignment")
		}
		c.nick = value
	}
}

func realName(value string) option {
	return func(c *config) {
		if c.realname != "" {
			log.Fatalf("Repeated RealName assignment")
		}
		c.realname = value
	}
}

func ident(value string) option {
	return func(c *config) {
		if c.ident != "" {
			log.Fatalf("Repeated Ident assignment")
		}
		c.ident = value
	}
}

func password(value string) option {
	return func(c *config) {
		if c.password != "" {
			log.Fatalf("Repeated Password assignment")
		}
		c.password = value
	}
}

func weatherToken(value string) option {
	return func(c *config) {
		if c.weatherToken != "" {
			log.Fatalf("Repeated WeatherToken assignment")
		}
		c.weatherToken = value
	}
}

func server(value string) option {
	return func(c *config) {
		if c.server != "" {
			log.Fatalf("Repeated Server assignment")
		}
		c.server = value
	}
}

func channels(value string) option {
	return func(c *config) {
		if len(c.channels) != 0 {
			log.Fatalf("Repeated Channel assignment")
		}
		c.channels = strings.Split(value, ",")
	}
}

func dbname(db string) option {
	return func(c *config) {
		if c.dbname != "" {
			log.Fatalf("Repeated DBName assignment")
		}
		c.dbname = db
	}
}

func userAgent(value string) option {
	return func(c *config) {
		if c.userAgent != "" {
			log.Fatalf("Repeated UserAgent assignment")
		}
		c.userAgent = value
	}
}

func ignored(value string) option {
	return func(c *config) {
		if len(c.ignored) != 0 {
			log.Fatalf("Repeated Ignored assignment")
		}
		c.ignored = strings.Split(value, ",")
	}
}

func admins(value string) option {
	return func(c *config) {
		if len(c.admins) != 0 {
			log.Fatalf("Repeated Admins assignment")
		}
		c.admins = strings.Split(value, ",")
	}
}

func nickservPass(value string) option {
	return func(c *config) {
		if c.nickservPass != "" {
			log.Fatalf("Repeated NickServ assignment")
		}
		c.nickservPass = value
	}
}

func pubFingerprint(value string) option {
	return func(c *config) {
		if c.pubFingerprint != "" {
			log.Fatalf("Repeated PubFingerprint assignment")
		}
		c.pubFingerprint = value
	}
}

func timeout(value string) option {
	return func(c *config) {
		if c.timeout != time.Duration(0) {
			log.Fatalf("Repeated timeout assignment")
		}
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			log.Fatalf("%q is not valid unsigned integer for Timeout", value)
		}
		c.timeout = time.Duration(n) * time.Second
	}
}
