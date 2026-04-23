package main

import (
	"fmt"
	"os"

	"euphoria.leet.nu/heim/proto/snowflake"
)

func main() {
	sf, err := snowflake.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	fmt.Println(sf.String())
}
