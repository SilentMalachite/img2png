package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/SilentMalachite/img2png/internal/archiver"
	"github.com/SilentMalachite/img2png/internal/converter"
	"github.com/SilentMalachite/img2png/internal/gui"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		// No args → GUI mode.
		if err := gui.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return
	}

	if len(args) == 1 {
		switch args[0] {
		case "-h", "--help":
			printHelp()
			return
		}
		if err := run(args[0]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			pause()
			os.Exit(1)
		}
		pause()
		return
	}

	fmt.Fprintln(os.Stderr, "Usage: img2png <file or folder>")
	fmt.Fprintln(os.Stderr, "Run img2png without arguments to open the GUI.")
	pause()
	os.Exit(1)
}

func printHelp() {
	fmt.Println(`img2png — convert images to PNG.

Usage:
  img2png                    Open the GUI.
  img2png <file>             Convert a single file; PNG is written next to it.
  img2png <folder>           Convert every supported image in <folder>;
                             a <folder>.zip is written next to it.
  img2png -h | --help        Show this help.

Supported inputs: .tif .tiff .jpg .jpeg .webp .bmp .gif`)
}

func run(path string) error {
	path = filepath.Clean(path)

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access: %s", path)
	}

	if info.IsDir() {
		return runDir(path)
	}
	return runFile(path)
}

func runFile(path string) error {
	if !supportedExt(filepath.Ext(path)) {
		return fmt.Errorf("unsupported format: %s", filepath.Base(path))
	}

	outDir := filepath.Dir(path)
	if err := checkWritable(outDir); err != nil {
		return err
	}

	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	outName := base + ".png"

	outPath, err := converter.ConvertFile(path, outDir, outName)
	if err != nil {
		return err
	}

	fmt.Println("Done:", filepath.Base(outPath))
	return nil
}

func runDir(dirPath string) error {
	parentDir := filepath.Dir(dirPath)
	if err := checkWritable(parentDir); err != nil {
		return err
	}

	var files []string
	_ = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && supportedExt(filepath.Ext(path)) {
			files = append(files, path)
		}
		return nil
	})

	if len(files) == 0 {
		return fmt.Errorf("no supported image files found")
	}

	tmpDir, err := os.MkdirTemp("", "img2png_*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	seen := make(map[string]int)
	var pngFiles []string
	for _, src := range files {
		base := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src)) + ".png"
		outName := archiver.DedupeFilename(base, seen)
		seen[base]++

		pngPath, err := converter.ConvertFile(src, tmpDir, outName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skipped: %s (unsupported or broken)\n", filepath.Base(src))
			continue
		}
		pngFiles = append(pngFiles, pngPath)
	}

	if len(pngFiles) == 0 {
		return fmt.Errorf("all files failed to convert")
	}

	dirName := filepath.Base(dirPath)
	zipPath := filepath.Join(parentDir, dirName+".zip")

	if err := archiver.Archive(pngFiles, zipPath); err != nil {
		return fmt.Errorf("failed to create ZIP: %w", err)
	}

	fmt.Println("Done:", filepath.Base(zipPath))
	return nil
}

func supportedExt(ext string) bool {
	return converter.IsSupportedExt(ext, false)
}

func checkWritable(dir string) error {
	tmp, err := os.CreateTemp(dir, ".img2png_check_*")
	if err != nil {
		return fmt.Errorf("cannot write to directory: %s", dir)
	}
	tmp.Close()
	os.Remove(tmp.Name())
	return nil
}

func pause() {
	fmt.Fprint(os.Stderr, "Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
