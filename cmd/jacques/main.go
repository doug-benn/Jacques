package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/doug-benn/Jacques/internal/processor"
)

func parseFlags(args []string) (input, output string, version, experimentalFolding bool, err error) {
	fs := flag.NewFlagSet("jacques", flag.ContinueOnError)
	fs.SetOutput(&strings.Builder{})

	fs.StringVar(&input, "i", "-", "Input file (default: stdin)")
	fs.StringVar(&input, "input", "-", "Input file (default: stdin)")
	fs.StringVar(&output, "o", "-", "Output file (default: stdout)")
	fs.StringVar(&output, "output", "-", "Output file (default: stdout)")
	fs.BoolVar(&version, "version", false, "Print version information")
	fs.BoolVar(&version, "v", false, "Print version information")
	fs.BoolVar(&experimentalFolding, "experimental-folding", false, "Enable features not covered by E2E tests")

	err = fs.Parse(args)
	if err != nil {
		return "", "", false, false, err
	}

	return input, output, version, experimentalFolding, nil
}

func run(input, output string, experimentalFolding bool, in io.Reader, out io.Writer) error {
	var r io.Reader

	if input == "-" && in != nil {
		r = in
	} else if input == "-" {
		stat, err := os.Stdin.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) != 0 {
			return fmt.Errorf("no input specified")
		}
		r = os.Stdin
	} else {
		f, err := os.Open(input)
		if err != nil {
			return fmt.Errorf("opening input file %q: %w", input, err)
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	var w io.Writer
	if output == "-" && out != nil {
		w = out
	} else if output == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("creating output file %q: %w", output, err)
		}
		defer func() { _ = f.Close() }()
		w = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return fmt.Errorf("no input specified")
	}

	opts := &processor.Options{ExperimentalFolding: experimentalFolding}
	result := processor.Process(string(data), opts)

	_, err = fmt.Fprintln(w, result)
	if err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

func main() {
	input, output, version, experimentalFolding, err := parseFlags(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			printUsage()
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		printUsage()
		os.Exit(1)
	}

	if version {
		fmt.Printf("Jacques %s\n", Version)
		os.Exit(0)
	}

	if err := run(input, output, experimentalFolding, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Jacques %s\n\n", Version)
	fmt.Fprintf(os.Stderr, "Usage: jacques [options]\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}
