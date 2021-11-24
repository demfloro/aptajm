package main

import (
	"golang.org/x/sys/unix"
	"log"
)

func init() {
	unveils := map[string]string{
		"/dev/log":         "rw",
		"/etc/resolv.conf": "r",
		"/etc/ssl":         "r",
		"/var/gobot":       "rwc",
		"/var/log/gobot":   "rwc",
	}
	for path, rights := range unveils {
		if err := unix.Unveil(path, rights); err != nil {
			log.Fatalf("Unveil failed for %q:%q: %q", path, rights, err)
		}
	}

	if err := unix.Pledge("stdio rpath wpath cpath flock inet unix dns", ""); err != nil {
		log.Fatalf("Pledge failed: %q", err)
	}
}
