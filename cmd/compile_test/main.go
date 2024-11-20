package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"

	"github.com/bobertlo/gmars"
)

func listFiles(dirpath string) ([]string, error) {
	output := make([]string, 0)

	fileSystem := os.DirFS(dirpath)

	err := fs.WalkDir(fileSystem, ".", func(filepath string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(path.Ext(filepath))
		if ext != ".red" {
			return nil
		}

		output = append(output, filepath)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

func main() {
	// legacyFlag := flag.Bool("8", false, "compile in ICWS88 mode")
	presetFlag := flag.String("preset", "nop94", "preset to use when compiling files")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: compile_test [-preset name] <code_dir> <compiled_dir>\n")
		os.Exit(1)
	}

	config, err := gmars.PresetConfig(*presetFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading preset: %s\n", err)
		os.Exit(1)
	}

	fileList, err := listFiles(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input dir '%s': %s\n", args[0], err)
	}

	compileErrors := 0
	compileSuccess := 0
	matches := 0
	mismatches := 0
	for _, file := range fileList {
		inPath := path.Join(args[0], file)
		expectedPath := path.Join(args[1], file)

		fIn, err := os.Open(inPath)
		if err != nil {
			fmt.Printf("error reading file '%s': %s\n", inPath, err)
			continue
		}
		defer fIn.Close()

		fExpected, err := os.Open(expectedPath)
		if err != nil {
			fmt.Printf("error reading file '%s': %s\n", inPath, err)
			continue
		}
		defer fExpected.Close()

		in, err := gmars.CompileWarrior(fIn, config)
		if err != nil {
			fmt.Printf("error compiling warrior '%s': %s\n", inPath, err)
			compileErrors++
			continue
		}

		expected, err := gmars.ParseLoadFile(fExpected, config)
		if err != nil {
			fmt.Printf("error loading warrior '%s': %s\n", expectedPath, err)
			compileErrors++
			continue
		}

		compileSuccess++

		if len(in.Code) != len(expected.Code) {
			fmt.Printf("%s: length mismatch: %d != %d", inPath, len(in.Code), len(expected.Code))
			mismatches++
			continue
		}

		instructionsMatch := true
		for i, inst := range in.Code {
			if expected.Code[i] != inst {
				fmt.Printf("%s: instruction mismatch: '%s' != '%s'\n", inPath, inst, expected.Code[i])
				instructionsMatch = false
			}
		}

		if instructionsMatch {
			matches++
		} else {
			mismatches++
		}
	}

	fmt.Println(compileErrors, "errors")
	fmt.Println(compileSuccess, "successfully compiled")
	fmt.Println(mismatches, "mismatches")
	fmt.Println(matches, "matches")
}
