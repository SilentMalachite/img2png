// Generates assets/icon.png — a 512×512 solid-color placeholder.
// Run once: go run ./cmd/genicon
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func main() {
	const size = 512
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	bg := color.RGBA{R: 0x2a, G: 0x68, B: 0xa8, A: 0xff}
	fg := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, bg)
		}
	}
	for y := 192; y < 320; y++ {
		for x := 192; x < 320; x++ {
			img.Set(x, y, fg)
		}
	}

	if err := os.MkdirAll("assets", 0o755); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("assets/icon.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/icon.png")
}
