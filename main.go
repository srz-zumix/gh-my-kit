/*
Copyright © 2025 srz_zumix
*/
package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/srz-zumix/gh-my-kit/cmd"
)

func main() {
	// Load .env file if present, unless GH_MY_KIT_NO_DOTENV is set.
	if os.Getenv("GH_MY_KIT_NO_DOTENV") == "" {
		if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
			// Log non-NotExist errors to help diagnose configuration issues
			fmt.Fprintln(os.Stderr, "failed to load .env file:", err)
		}
	}
	cmd.Execute()
}
