package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"

	"github.com/alecthomas/kong"
)

const (
	white = 0
	black = 255
)

var cli struct {
	File   string `arg name:"path" help:"File to be thresholded." type:"existingfile"`
	Output string `help:"Place the output into <file>." short:"o" default:"output.png" type:"path"`
}

func main() {
	_ = kong.Parse(&cli, kong.Name("threshold"))
	if err := run(cli.File, cli.Output); err != nil {
		log.Fatal(err)
	}
}

func run(src, dst string) error {
	img, err := loadImage(src)
	if err != nil {
		return err
	}

	gImg := toGray(img)
	gImg = threshold(gImg, otsu(gImg), white, black)
	if err := saveImage(dst, gImg); err != nil {
		return err
	}

	return nil
}

func loadImage(n string) (image.Image, error) {
	f, err := os.Open(n)
	defer f.Close()
	if err != nil {
		return nil, fmt.Errorf("could not open file: %s", n)
	}
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("could not decode: %s", n)
	}
	return img, nil
}

func saveImage(n string, img image.Image) error {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(n, buf.Bytes(), 0644)
}

func toGray(img image.Image) *image.Gray {
	rct := img.Bounds()
	res := image.NewGray(rct)
	for y := rct.Min.Y; y < rct.Max.Y; y++ {
		for x := rct.Min.X; x < rct.Max.X; x++ {
			c := color.GrayModel.Convert(img.At(x, y))
			gray, _ := c.(color.Gray)
			res.Set(x, y, gray)
		}
	}
	return res
}

func threshold(img *image.Gray, th, bg, fg uint8) *image.Gray {
	res := image.NewGray(img.Bounds())
	for i := 0; i < len(res.Pix); i++ {
		if img.Pix[i] > th {
			res.Pix[i] = fg
		} else {
			res.Pix[i] = bg
		}
	}
	return res
}

func otsu(img *image.Gray) uint8 {
	hist := histogram(img)
	sum := 0
	for i, v := range hist {
		sum += i * v
	}
	na, nb := 0, len(img.Pix)
	sa, sb := 0, sum
	ms, th := 0.0, uint8(0)
	for t := 0; t < 256; t++ {
		na += hist[t]
		nb -= hist[t]
		if na == 0 {
			continue
		}
		if nb == 0 {
			break
		}
		sa += t * hist[t]
		sb = sum - sa
		ma := float64(sa) / float64(na)
		mb := float64(sb) / float64(nb)
		s := float64(na*nb) * (ma - mb) * (ma - mb)
		if s > ms {
			ms = s
			th = uint8(t)
		}
	}
	return th
}

func histogram(img *image.Gray) []int {
	hist := make([]int, 256)
	for i := 0; i < len(img.Pix); i++ {
		hist[img.Pix[i]]++
	}
	return hist
}
