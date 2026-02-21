package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/doug-benn/Jacques/internal/processor"
)

func main() {
	var input string
	var output string
	var version bool
	flag.StringVar(&input, "i", "-", "Input file (default: stdin)")
	flag.StringVar(&input, "input", "-", "Input file (default: stdin)")
	flag.StringVar(&output, "o", "-", "Output file (default: stdout)")
	flag.StringVar(&output, "output", "-", "Output file (default: stdout)")
	flag.BoolVar(&version, "version", false, "Print version information")
	flag.BoolVar(&version, "v", false, "Print version information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Jacques %s\n\n", Version)
		fmt.Fprintf(os.Stderr, "Usage: jacques [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if version {
		fmt.Printf("Jacques %s\n", Version)
		os.Exit(0)
	}

	var in io.Reader
	var out io.Writer

	if input == "-" {
		// Check if stdin is a terminal (interactive) - if so, show error since
		// there's no way to provide input interactively
		stat, err := os.Stdin.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) != 0 {
			fmt.Fprintf(os.Stderr, "Error: no input specified\n")
			flag.Usage()
			os.Exit(1)
		}
		in = os.Stdin
	} else {
		f, err := os.Open(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", fmt.Errorf("opening input file %q: %w", input, err))
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()
		in = f
	}

	if output == "-" {
		out = os.Stdout
	} else {
		f, err := os.Create(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", fmt.Errorf("creating output file %q: %w", output, err))
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()
		out = f
	}

	data, err := io.ReadAll(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", fmt.Errorf("reading input: %w", err))
		os.Exit(1)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no input specified\n")
		flag.Usage()
		os.Exit(1)
	}

	result := processor.Process(string(data))

	_, err = fmt.Fprintln(out, result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", fmt.Errorf("writing output: %w", err))
		os.Exit(1)
	}
}
