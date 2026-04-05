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
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: img2png <file or folder>")
		pause()
		os.Exit(1)
	}

	if err := run(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		pause()
		os.Exit(1)
	}

	pause()
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
	// ZIP は dirPath と同階層に生成
	parentDir := filepath.Dir(dirPath)
	if err := checkWritable(parentDir); err != nil {
		return err
	}

	// 対象ファイルを再帰的に収集
	var files []string
	_ = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // スキップして続行
		}
		if !d.IsDir() && supportedExt(filepath.Ext(path)) {
			files = append(files, path)
		}
		return nil
	})

	if len(files) == 0 {
		return fmt.Errorf("no supported image files found")
	}

	// 中間 PNG 用一時ディレクトリ
	tmpDir, err := os.MkdirTemp("", "img2png_*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// 変換（重複ファイル名を解決しながら）
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

	// ZIP 生成
	dirName := filepath.Base(dirPath)
	zipPath := filepath.Join(parentDir, dirName+".zip")

	if err := archiver.Archive(pngFiles, zipPath); err != nil {
		return fmt.Errorf("failed to create ZIP: %w", err)
	}

	fmt.Println("Done:", filepath.Base(zipPath))
	return nil
}

// supportedExt reports whether ext is a supported image extension.
func supportedExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".tif", ".tiff", ".jpg", ".jpeg", ".webp", ".bmp", ".gif":
		return true
	}
	return false
}

// checkWritable verifies write permission by creating and removing a temp file.
func checkWritable(dir string) error {
	tmp, err := os.CreateTemp(dir, ".img2png_check_*")
	if err != nil {
		return fmt.Errorf("cannot write to directory: %s", dir)
	}
	tmp.Close()
	os.Remove(tmp.Name())
	return nil
}

// pause はユーザーの Enter キー入力を待つ（ドラッグ&ドロップ利用者向け）。
func pause() {
	fmt.Fprint(os.Stderr, "Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
