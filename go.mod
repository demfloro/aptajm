module gitea.demsh.org/demsh/aptajm

go 1.17

require (
	gitea.demsh.org/demsh/ircfw v0.1.0
	github.com/mattn/go-sqlite3 v1.14.15
	golang.org/x/net v0.0.0-20220909164309-bea034e7d591
	golang.org/x/sys v0.0.0-20220915200043-7b5979e65e41
	golang.org/x/text v0.3.7
)

require gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637

replace gitea.demsh.org/demsh/ircfw => ../ircfw
