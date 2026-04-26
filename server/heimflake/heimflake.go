package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"euphoria.leet.nu/heim/proto/snowflake"
)

var (
	localMode = flag.Bool("l", false, "Convert timestamps into local timezone")
)

func formatTime(t time.Time) string {
	if *localMode {
		t = t.In(time.Local)
	}
	return t.Format("2006-01-02 15:04:05.000 Z0700")
}

func reportError(err error, fatal bool) bool {
	if err == nil {
		return false
	}
	fmt.Fprintln(os.Stderr, "ERROR:", err)
	if fatal {
		os.Exit(1)
	}
	return true
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		sf, err := snowflake.New()
		reportError(err, true)
		fmt.Println(sf.String())
		return
	}
	for _, s := range args {
		sf, err := snowflake.NewFromString(s)
		if reportError(err, false) {
			continue
		}
		fmt.Printf("%s: %s\n", sf, formatTime(sf.Time()))
	}
}
