module gitea.demsh.org/demsh/aptajm

go 1.17

require (
	gitea.demsh.org/demsh/ircfw v0.1.0
	github.com/mattn/go-sqlite3 v1.14.19
	golang.org/x/net v0.20.0
	golang.org/x/sys v0.16.0
	golang.org/x/text v0.14.0
)

require gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637

replace gitea.demsh.org/demsh/ircfw => ../ircfw
