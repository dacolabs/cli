package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dacolabs/daco/cmd/cli/app"
)

func main() {
	if err := app.Run(context.Background(), os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}