package presence

import (
	"net"
	"net/http"
	"sync"

	"github.com/euphoria-io/scope"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"euphoria.leet.nu/heim/proto/logging"
)

func Serve(ctx scope.Context, addr string) {
	http.Handle("/metrics", promhttp.Handler())

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logging.Logger(ctx).Printf("http[%s]: %s\n", addr, err)
		ctx.Terminate(err)
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

	if err := http.Serve(listener, nil); err != nil {
		logging.Logger(ctx).Printf("http[%s]: %s\n", addr, err)
		ctx.Terminate(err)
	}

	closeListener()
	ctx.WaitGroup().Done()
}
