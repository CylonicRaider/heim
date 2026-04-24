package main

import (
	"fmt"
	"os"

	"euphoria.leet.nu/heim/proto/snowflake"
)

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
	if len(os.Args) <= 1 {
		sf, err := snowflake.New()
		reportError(err, true)
		fmt.Println(sf.String())
		return
	}
	for _, s := range os.Args[1:] {
		sf, err := snowflake.NewFromString(s)
		if reportError(err, false) {
			continue
		}
		fmt.Printf("%s: %s\n", sf, sf.Time())
	}
}
