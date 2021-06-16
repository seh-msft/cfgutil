// Copyright (c) 2021, Microsoft Corporation, Sean Hinchee
// Licensed under the MIT License.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/seh-msft/cfg"
	"github.com/seh-msft/openapi"
)

var (
	mkMode     = flag.Bool("mk", false, "Generate a new cfg file (default)")
	jsonMode   = flag.Bool("json", false, "Convert a cfg file to JSON")
	cfgFile    = flag.String("cfg", "", "Input .cfg file (json)")
	apiFile    = flag.String("api", "", "Input .json OpenAPI specification file (mk)")
	outFile    = flag.String("o", "", "Output file")
	strict     = flag.Bool("strict", false, "Generate a strict cfg allowlisting explicit path:title combinations (mk)")
	everything = flag.Bool("all", false, "Include every parameter in the output (mk)")
	useSingle  = flag.Bool("single", false, "Force usage of single quoting")
	noAPI      = flag.Bool("minimal", false, "If not in strict mode, do not emit exclusivity parameters (mk)")
	cautious   = flag.Bool("cautious", false, "")
	quote      = '"'
)

// Cfg utility for generating cfg files from openapi specifications.
func main() {
	flag.Parse()
	args := flag.Args()

	// Output file handling
	var out *bufio.Writer = bufio.NewWriter(os.Stdout)
	if len(*outFile) > 0 {
		f, err := os.Create(*outFile)
		if err != nil {
			fatal("err: could not open output file →", err)
		}
		out = bufio.NewWriter(f)
		defer f.Close()
	}
	defer out.Flush()

	if *jsonMode && !*mkMode {
		toJSON(args, out)
		return
	}

	mk(args, out)
}

// Convert a cfg file to valid JSON
func toJSON(args []string, out *bufio.Writer) {
	if (len(args) > 0 && len(*cfgFile) > 0) || (len(args) <= 0 && *cfgFile == "") {
		fatal("err: one of -cfg or an argument file must be provided")
	}

	var path string = *cfgFile
	if len(path) < 1 {
		path = args[0]
	}

	f, err := os.Open(path)
	if err != nil {
		fatal("err: could not open file →", err)
	}

	if *useSingle {
		cfg.Quoting = cfg.Single
	} else {
		cfg.Quoting = cfg.Double
	}
	c, err := cfg.Load(f)
	if err != nil {
		fatal("err: could not cfg parse file →", err)
	}

	// Encode to JSON
	var buf strings.Builder
	c.Emit(&buf)
	enc := json.NewEncoder(out)
	err = enc.Encode(buf.String())
	if err != nil {
		fatal("err: could not encode to JSON →", err)
	}
}

// Generate a new cfg file for one or more OpenAPI specifications
// Build a valid .cfg for all required identifiers in an OpenAPI specification
// One API can be specified via -i or a variable number can be passed as arguments
func mk(args []string, out *bufio.Writer) {
	if (len(args) > 0 && len(*apiFile) > 0) || (len(args) <= 0 && *apiFile == "") {
		fatal("err: one of -api or a list of argument specification files must be provided")
	}

	// Input file handling
	var apis []openapi.API

	if len(*apiFile) > 0 {
		// One file
		apis = append(apis, f2api(*apiFile))

	} else {
		// Many files as arguments
		for _, file := range args {
			apis = append(apis, f2api(file))
		}
	}

	if *useSingle {
		quote = '\''
	}

	var do func(api openapi.API, out io.Writer) = doLoose
	if *strict {
		do = doStrict
	}

	for _, api := range apis {
		do(api, out)
	}
}

func doLoose(api openapi.API, out io.Writer) {
	title := clean(api.Info.Title)
	const tmpl = `%s=
`
	var constraints = `	disallow path=%c.*%c title=%c.*%c
	permit title=%s
`
	if !*cautious {
		constraints = `	disallow path=.* title=.*
	permit title=%s
`
	}

	fmt.Fprintf(out, "# Identifiers for the API %s:\n\n", title)

	names := make(map[string]string)
	for _, methods := range api.Paths {
		for _, method := range methods {
			for _, parameter := range method.Parameters {
				if !parameter.Required && !*everything {
					// Skip parameters that aren't required
					continue
				}

				names[clean(parameter.Name)] = ""
			}
		}
	}

	for name := range names {
		// Emit identifiers
		fmt.Fprintf(out, tmpl, name)
		if !*noAPI {
			if *cautious {
				fmt.Fprintf(out, constraints, quote, quote, quote, quote, title)
			} else {
				fmt.Fprintf(out, constraints, title)
			}
		}

		fmt.Fprintf(out, "\n")
	}
}

func doStrict(api openapi.API, out io.Writer) {
	title := clean(api.Info.Title)

	const tmpl = `%s=
	disallow path=%c.*%c title=%c.*%c
	permit path=%s title=%s

`

	fmt.Fprintf(out, "# Identifiers for the API %s:\n\n", title)

	for path, methods := range api.Paths {
		path = clean(path)
		for _, method := range methods {
			for _, parameter := range method.Parameters {
				if !parameter.Required && !*everything {
					// Skip parameters that aren't required
					continue
				}

				name := clean(parameter.Name)

				fmt.Fprintf(out, tmpl, name, quote, quote, quote, quote, path, title)
			}
		}
	}
}

// Double quote escape quote literals, if any
// Quote wrap string
func clean(s string) string {

	out := strings.ReplaceAll(s, string(quote), string(quote)+string(quote))
	if *cautious {
		return string(quote) + out + string(quote)
	}

	for _, rune := range out {
		if unicode.IsSpace(rune) {
			return string(quote) + out + string(quote)
		}
	}

	return out
}

// Open an API
func f2api(path string) openapi.API {
	f, err := os.Open(path)
	if err != nil {
		fatal("err: could not open API file →", err)
	}
	defer f.Close()

	api, err := openapi.Parse(f)
	if err != nil {
		fatal("err: could not parse API →", err)
	}

	return api
}

// Fatal - end program with an error message and newline
func fatal(s ...interface{}) {
	fmt.Fprintln(os.Stderr, s...)
	os.Exit(1)
}
