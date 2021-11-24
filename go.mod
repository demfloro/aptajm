module gitea.demsh.org/demsh/aptajm

go 1.17

require (
	gitea.demsh.org/demsh/ircfw v0.1.0
	github.com/mattn/go-sqlite3 v1.14.9
	golang.org/x/net v0.0.0-20211123203042-d83791d6bcd9
	golang.org/x/sys v0.0.0-20211123173158-ef496fb156ab
	golang.org/x/text v0.3.7
)

require gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637

replace gitea.demsh.org/demsh/ircfw => ../ircfw
