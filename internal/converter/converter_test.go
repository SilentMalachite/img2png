package converter

import (
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/tiff"
)

// createTestImage generates a 2x2 RGBA image for testing.
func createTestImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{G: 255, A: 255})
	img.Set(0, 1, color.RGBA{B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	return img
}

func writeTestJPEG(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, createTestImage(), nil); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeTestGIF(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img := createTestImage()
	palettedImg := image.NewPaletted(img.Bounds(), color.Palette{
		color.RGBA{R: 255, A: 255},
		color.RGBA{G: 255, A: 255},
		color.RGBA{B: 255, A: 255},
		color.RGBA{R: 255, G: 255, B: 255, A: 255},
	})
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			palettedImg.Set(x, y, img.At(x, y))
		}
	}
	if err := gif.Encode(f, palettedImg, nil); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeTestTIFF(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := tiff.Encode(f, createTestImage(), nil); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeTestBMP(t *testing.T, dir, name string) string {
	t.Helper()
	// Minimal 1x1 24-bit BMP (white pixel)
	bmp := []byte{
		// File header (14 bytes)
		0x42, 0x4D, // "BM"
		0x3A, 0x00, 0x00, 0x00, // file size: 58
		0x00, 0x00, 0x00, 0x00, // reserved
		0x36, 0x00, 0x00, 0x00, // offset to pixel data: 54
		// DIB header (40 bytes)
		0x28, 0x00, 0x00, 0x00, // header size: 40
		0x01, 0x00, 0x00, 0x00, // width: 1
		0x01, 0x00, 0x00, 0x00, // height: 1
		0x01, 0x00, // planes: 1
		0x18, 0x00, // bpp: 24
		0x00, 0x00, 0x00, 0x00, // compression: none
		0x04, 0x00, 0x00, 0x00, // image size: 4 (1 pixel + 1 byte padding)
		0x13, 0x0B, 0x00, 0x00, // x ppm
		0x13, 0x0B, 0x00, 0x00, // y ppm
		0x00, 0x00, 0x00, 0x00, // colors
		0x00, 0x00, 0x00, 0x00, // important colors
		// Pixel data (4 bytes: BGR + 1 byte row padding)
		0xFF, 0xFF, 0xFF, 0x00,
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, bmp, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeTestWebP(t *testing.T, dir, name string) string {
	t.Helper()
	// Minimal 1x1 lossy WebP (VP8 format, generated via cwebp)
	webp := []byte{
		0x52, 0x49, 0x46, 0x46, 0x24, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50,
		0x56, 0x50, 0x38, 0x20, 0x18, 0x00, 0x00, 0x00, 0x30, 0x01, 0x00, 0x9d,
		0x01, 0x2a, 0x01, 0x00, 0x01, 0x00, 0x02, 0x00, 0x34, 0x25, 0xa4, 0x00,
		0x03, 0x70, 0x00, 0xfe, 0xfb, 0x94, 0x00, 0x00,
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, webp, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestConvertFile_JPEG(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()
	src := writeTestJPEG(t, srcDir, "test.jpg")

	got, err := ConvertFile(src, outDir, "test.png")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "test.png" {
		t.Errorf("output name: got %q, want %q", filepath.Base(got), "test.png")
	}
	verifyPNG(t, got)
}

func TestConvertFile_GIF(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()
	src := writeTestGIF(t, srcDir, "anim.gif")

	got, err := ConvertFile(src, outDir, "anim.png")
	if err != nil {
		t.Fatal(err)
	}
	verifyPNG(t, got)
}

func TestConvertFile_TIFF(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()
	src := writeTestTIFF(t, srcDir, "scan.tiff")

	got, err := ConvertFile(src, outDir, "scan.png")
	if err != nil {
		t.Fatal(err)
	}
	verifyPNG(t, got)
}

func TestConvertFile_BMP(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()
	src := writeTestBMP(t, srcDir, "icon.bmp")

	got, err := ConvertFile(src, outDir, "icon.png")
	if err != nil {
		t.Fatal(err)
	}
	verifyPNG(t, got)
}

func TestConvertFile_WebP(t *testing.T) {
	srcDir := t.TempDir()
	outDir := t.TempDir()
	src := writeTestWebP(t, srcDir, "photo.webp")

	got, err := ConvertFile(src, outDir, "photo.png")
	if err != nil {
		t.Fatal(err)
	}
	verifyPNG(t, got)
}

func TestConvertFile_InvalidFile(t *testing.T) {
	srcDir := t.TempDir()
	path := filepath.Join(srcDir, "garbage.jpg")
	if err := os.WriteFile(path, []byte("not an image"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ConvertFile(path, t.TempDir(), "garbage.png")
	if err == nil {
		t.Error("expected error for invalid file, got nil")
	}
}

func TestConvertFile_MissingFile(t *testing.T) {
	_, err := ConvertFile("/nonexistent/path.jpg", t.TempDir(), "out.png")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestIsSupportedExt(t *testing.T) {
	cases := []struct {
		ext      string
		allowPNG bool
		want     bool
	}{
		{".jpg", false, true}, {".JPEG", true, true},
		{".png", false, false}, {".png", true, true},
		{".PNG", true, true},
		{".txt", true, false}, {"", true, false},
	}
	for _, c := range cases {
		if got := IsSupportedExt(c.ext, c.allowPNG); got != c.want {
			t.Errorf("IsSupportedExt(%q, %v) = %v, want %v", c.ext, c.allowPNG, got, c.want)
		}
	}
}

// verifyPNG checks that the file at path is a valid PNG.
func verifyPNG(t *testing.T, path string) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("cannot open output: %v", err)
	}
	defer f.Close()

	_, err = png.Decode(f)
	if err != nil {
		t.Fatalf("output is not a valid PNG: %v", err)
	}
}
