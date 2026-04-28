package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"euphoria.leet.nu/heim/proto/snowflake"
)

var (
	localMode   = flag.Bool("l", false, "Convert timestamps into local timezone")
	verboseMode = flag.Bool("v", false, "Show more details")
)

func formatTime(t time.Time) string {
	if *localMode {
		t = t.In(time.Local)
	}
	return t.Format("2006-01-02 15:04:05.000 Z0700")
}

func formatDetails(sf snowflake.Snowflake) string {
	parts := sf.Split()
	return fmt.Sprintf(`
snowflake: %s (%d)
time     : %s (%d)
worker   : %d
sequence : %d`[1:],
		sf, uint64(sf),
		formatTime(parts.Time), uint64(time.Duration(parts.Time.UnixNano())/time.Millisecond),
		parts.WorkerID, parts.Sequence)
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
		if *verboseMode {
			fmt.Println(formatDetails(sf))
		} else {
			fmt.Println(sf)
		}
		return
	}
	for i, s := range args {
		sf, err := snowflake.NewFromString(s)
		if reportError(err, false) {
			continue
		}
		if *verboseMode {
			if i != 0 {
				fmt.Println()
			}
			fmt.Println(formatDetails(sf))
		} else {
			fmt.Printf("%s: %s\n", sf, formatTime(sf.Time()))
		}
	}
}
