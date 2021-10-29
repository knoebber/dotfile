package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/niklasfasching/go-org/org"
)

// Converts org files in directory to html.
func main() {
	inDir := flag.String("in", "", "directory with .org files")
	outDir := flag.String("out", "", "directory to output html files")
	flag.Parse()

	if err := convert(*inDir, *outDir); err != nil {
		fmt.Printf("failed to generate docs: %s\n", err)
	}
}

func convert(inDir, outDir string) error {
	orgFiles, err := filepath.Glob(filepath.Join(inDir, "*.org"))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	for _, path := range orgFiles {
		outputName := strings.Split(filepath.Base(path), ".")[0] + ".html"
		err = render(path, filepath.Join(outDir, outputName))
		if err != nil {
			return err
		}
	}

	return nil
}

func render(inPath, outPath string) error {
	bs, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}

	d := org.New().Parse(bytes.NewReader(bs), inPath)

	output, err := d.Write(org.NewHTMLWriter())
	if err != nil {
		return err
	}

	output = "<main>\n" + output + "</main>\n"

	if err := os.WriteFile(outPath, []byte(output), 0644); err != nil {
		return err
	}
	fmt.Println("Wrote", outPath)

	return nil
}
