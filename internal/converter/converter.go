package converter

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	// デコーダ登録（ブランクインポート）
	_ "image/gif"
	_ "image/jpeg"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// ConvertFile decodes the image at srcPath and writes a PNG to outDir/outName.
// Returns the full path of the created PNG file.
// For animated GIFs, only the first frame is converted.
func ConvertFile(srcPath, outDir, outName string) (string, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", filepath.Base(srcPath), err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", fmt.Errorf("decode %s: %w", filepath.Base(srcPath), err)
	}

	outPath := filepath.Join(outDir, outName)
	out, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", outName, err)
	}

	if err := png.Encode(out, img); err != nil {
		out.Close()
		os.Remove(outPath)
		return "", fmt.Errorf("encode %s: %w", outName, err)
	}

	if err := out.Close(); err != nil {
		os.Remove(outPath)
		return "", fmt.Errorf("write %s: %w", outName, err)
	}

	return outPath, nil
}
