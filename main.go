package main

import (
	"fmt"
	"os"
)

func main() {
	if err := getMainFunction().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err.Error())
		os.Exit(1)
	}
}
