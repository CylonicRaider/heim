package worker

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"euphoria.leet.nu/heim/proto/logging"
	"github.com/euphoria-io/scope"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Serve(ctx scope.Context, addr string) {
	http.Handle("/metrics", promhttp.Handler())

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("http[%s]: listen error: %s\n", addr, err)
		ctx.Terminate(err)
		return
	}

	closed := false
	m := sync.Mutex{}
	closeListener := func() {
		m.Lock()
		if !closed {
			listener.Close()
			closed = true
		}
		m.Unlock()
	}

	// Spin off goroutine to watch ctx and close listener if shutdown requested.
	go func() {
		<-ctx.Done()
		closeListener()
	}()

	logging.Logger(ctx).Printf("serving /metrics on %s", addr)
	if err := http.Serve(listener, nil); err != nil {
		fmt.Printf("http[%s]: %s\n", addr, err)
		ctx.Terminate(err)
	}

	closeListener()
	ctx.WaitGroup().Done()
}
