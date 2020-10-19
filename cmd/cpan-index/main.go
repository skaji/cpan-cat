package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/skaji/cpan-cat"
)

func run(ctx context.Context, args []string) error {
	if err := os.MkdirAll(cpan.BaseDir, 0777); err != nil {
		return err
	}
	f := cpan.NewFile("https://cpan.metacpan.org/modules/02packages.details.txt.gz")
	if err := f.Fetch(ctx); err != nil {
		return err
	}
	if len(args) > 1 && (args[1] == "-t" || args[1] == "-m") {
		t := f.ModTime()
		fmt.Printf(
			"02packages.details.txt: %s ago (%s)\n",
			time.Since(t).Round(time.Second).String(),
			t.Format(time.RFC3339),
		)
		return nil
	}
	return f.Cat(ctx, os.Stdout)
}

func main() {
	if err := run(context.Background(), os.Args); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
