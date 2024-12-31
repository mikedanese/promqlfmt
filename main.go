package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/mattn/go-isatty"
	"github.com/prometheus/prometheus/promql/parser"
)

const verbose = false

var (
	dryRun = flag.Bool("dry-run", false, "Print diffs instead of applying changes.")
	differ = flag.String("differ", "auto", "Diff command to use in dry run mode. Can be {auto, diff, gitdiff}. Auto will attempt to use gitdiff but fall back to diff.")
)

func main() {
	log.SetFlags(0)
	if verbose {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	}
	log.SetPrefix("promqlfmt: ")
	flag.Parse()

	var errs []error
	for _, path := range flag.Args() {
		if err := formatFile(path); err != nil {
			log.Print(err)
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		os.Exit(1)
	}
}

func formatFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", path, err)
	}
	p := parser.NewParser(string(data))

	expr, err := p.ParseExpr()
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}
	formatted := []byte(parser.Prettify(expr))
	if bytes.Equal(data, formatted) {
		// No formatting needed.
		return nil
	}
	if *dryRun {
		log.Printf("Diffing %s", path)
		if err := diff(path, formatted); err != nil {
			return fmt.Errorf("failed to diff %s: %v", path, err)
		}
	} else {
		log.Printf("Applying formatting to %s", path)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := writeFormatted(f, formatted); err != nil {
			return fmt.Errorf("failed to write %s: %v", path, err)
		}
	}
	return nil
}

var (
	haveDiff    = haveCommand("diff")
	haveGitDiff = haveCommand("git")
)

func diff(path string, formatted []byte) error {

	yr, yw, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %v", err)
	}
	defer yr.Close()

	go func() {
		defer yw.Close()
		if err := writeFormatted(yw, formatted); err != nil {
			log.Printf("Error while diffing: failed to write to pipe: %v", err)
		}
	}()

	if *differ == "auto" {
		if haveGitDiff {
			*differ = "gitdiff"
		} else if haveDiff {
			*differ = "diff"
		}
	}

	var cmd *exec.Cmd
	switch *differ {
	case "gitdiff":
		cmd = exec.Command("git", "diff", "--word-diff", path, "/dev/fd/3")
	case "diff":
		cmd = exec.Command("diff", "--unified=10", path, "/dev/fd/3")
		if isatty.IsTerminal(os.Stdout.Fd()) {
			cmd.Args = append(cmd.Args, "--color")
		}
	default:
		log.Fatalf("Unknown differ: %s", *differ)
	}

	cmd.ExtraFiles = []*os.File{yr}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return err
	}
	switch exitErr.ExitCode() {
	case 0, 1:
		return nil
	default:
		return err
	}
}

func writeFormatted(w io.Writer, formatted []byte) error {
	if _, err := w.Write(formatted); err != nil {
		return err
	}
	// Add a newline to the end of the file.
	if _, err := io.WriteString(w, "\n"); err != nil {
		return err
	}
	return nil
}

func haveCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
