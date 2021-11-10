package main

import (
	"context"
	"fmt"
	"runtime"

	"gitea.demsh.org/demsh/ircfw"
)

func handleStatus(ctx context.Context, bot *ircbot, msg ircfw.Msg) {
	var m runtime.MemStats
	if !msg.IsPrivate() {
		return
	}
	if !isAdmin(msg.Prefix()) {
		return
	}
	runtime.ReadMemStats(&m)
	msg.Reply(ctx, []string{fmt.Sprintf("goroutines: %d, heap: %d KB, GC runs: %d, runtime: %s",
		runtime.NumGoroutine(), m.HeapAlloc/1024, m.NumGC, runtime.Version())})
}
