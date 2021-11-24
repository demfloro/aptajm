package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

var (
	charmaps = map[string]*charmap.Charmap{
		"windows-1250": charmap.Windows1250,
		"windows-1251": charmap.Windows1251,
		"windows-1252": charmap.Windows1252,
		"windows-1253": charmap.Windows1253,
		"windows-1254": charmap.Windows1254,
		"windows-1255": charmap.Windows1255,
		"windows-1256": charmap.Windows1256,
		"windows-1257": charmap.Windows1257,
		"windows-1258": charmap.Windows1258,
		"cp1250":       charmap.Windows1250,
		"cp1251":       charmap.Windows1251,
		"cp1252":       charmap.Windows1252,
		"cp1253":       charmap.Windows1253,
		"cp1254":       charmap.Windows1254,
		"cp1255":       charmap.Windows1255,
		"cp1256":       charmap.Windows1256,
		"cp1257":       charmap.Windows1257,
		"cp1258":       charmap.Windows1258,
		"koi8r":        charmap.KOI8R,
		"koi8u":        charmap.KOI8U,
	}
)

type ConfigStruct struct {
	Nick, Password, Ident     string
	Realname, WeatherToken    string
	Server, DBName            string
	UserAgent, NickservPass   string
	PubFingerPrint            string
	Admins, Channels, Ignored []string
	Charset                   *charmap.Charmap
	Timeout                   time.Duration
}

func loadConfig(fname string) (*ConfigStruct, error) {
	if len(os.Args) < 2 {
		log.Fatalf("Not enough arguments, usage: %s configfile", os.Args[0])
	}
	file, err := os.Open(fname)
	if err != nil {
		log.Fatalf("Failed to open configuration file: %q", err)
	}
	defer file.Close()
	config, err := parseConfig(file)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func lookupCharset(name string) (result *charmap.Charmap) {
	if name == "" {
		return nil
	}
	result, ok := charmaps[name]
	if !ok {
		return nil
	}
	return result
}

func parseConfig(reader io.Reader) (c *ConfigStruct, err error) {
	scanner := bufio.NewScanner(reader)
	c = new(ConfigStruct)
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
		case "pubfingerprint":
			if c.PubFingerPrint != "" {
				log.Fatalf("Repeated Fingerprint assignment")
			}
			c.PubFingerPrint = v
		case "charset":
			if c.Charset != nil {
				log.Fatalf("Repeated Charset assignment")
			}
			c.Charset = lookupCharset(v)
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
		log.Fatalf("At least one channel in Channels must be specified")
	case c.Timeout == 0:
		c.Timeout = 10 * time.Second
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return c, nil

}
