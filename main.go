package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"log"
	"os"

	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"

	"github.com/nfnt/resize"
)

type Resources struct {
	Background  Images    `json:"background"`
	BgRgba      Rgba      `json:"bgRgba"`
	Like        Images    `json:"like"`
	Dislike     Images    `json:"dislike"`
	Output      OutImages `json:"output"`
	LikeIcon    []Images  `json:"likeIcon"`
	DislikeIcon []Images  `json:"dislikeIcon"`
}

type Rgba struct {
	Top    []uint8 `json:"top"`
	Bottom []uint8 `json:"bottom"`
}

type Images struct {
	Name   *string `json:"name"`
	File   *string `json:"file"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
}

type OutImages struct {
	File   *string `json:"file"`
	Format *string `json:"format"`
}

var (
	jsonPath = flag.String("j", "", "Json file path")
)

var usage = `Usage: %s [options...]
Options:
  -j  Json file path.

e.g.:
  goimage -j ./jsons/preference.json
`

func init() {
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, os.Args[0]))
	}

	if *jsonPath == "" {
		flag.Usage()

		os.Exit(1)
		return
	}

}

func main() {
	jsonByte, err := loadJSONFile(*jsonPath)
	if err != nil {
		log.Printf("fail to call loadJSONFile(): %v", err)
		return
	}

	var resources Resources
	err = json.Unmarshal(jsonByte, &resources)
	if err != nil {
		log.Printf("fail to call json.Unmarshal(): %v", err)
		return
	}

	createBgImage(&resources.Background, &resources.BgRgba)

	cImgs, lImgs, dlImgs := getImages(&resources)

	composeImage(&resources.Output, cImgs, lImgs, dlImgs)
}

func loadJSONFile(filePath string) ([]byte, error) {
	// Loading jsonfile
	if filePath == "" {
		return nil, errors.New("nothing JSON file")
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func createBgImage(bg *Images, rgba *Rgba) {
	img := image.NewRGBA(image.Rect(0, 0, bg.Width, bg.Height))

	for x := 0; x < bg.Width; x++ {
		for y := 0; y < 198; y++ {
			img.Set(x, y, color.RGBA{rgba.Top[0], rgba.Top[1], rgba.Top[2], rgba.Top[3]})
		}
		for y := 198; y < bg.Height; y++ {
			img.Set(x, y, color.RGBA{rgba.Bottom[0], rgba.Bottom[1], rgba.Bottom[2], rgba.Bottom[3]})
		}
	}

	// Save to out.png
	f, _ := os.OpenFile(*bg.File, os.O_WRONLY|os.O_CREATE, 0644)
	defer f.Close()
	png.Encode(f, img)
}

func getImages(r *Resources) ([]image.Image, []image.Image, []image.Image) {

	//1.open file
	bgFile, err := os.Open(*r.Background.File)
	if err != nil {
		log.Fatal(err)
	}

	likeFile, err := os.Open(*r.Like.File)
	if err != nil {
		log.Fatal(err)
	}

	dislikeFile, err := os.Open(*r.Dislike.File)
	if err != nil {
		log.Fatal(err)
	}

	//1.2 like
	var likeFiles []*os.File
	for _, v := range r.LikeIcon {
		file, err := os.Open(*v.File)
		if err != nil {
			log.Fatal(err)
		}
		likeFiles = append(likeFiles, file)
	}

	//1.3 dislike
	var dislikeFiles []*os.File
	for _, v := range r.DislikeIcon {
		file, err := os.Open(*v.File)
		if err != nil {
			log.Fatal(err)
		}
		dislikeFiles = append(dislikeFiles, file)
	}

	//2.decode background
	bgImg, _, err := image.Decode(bgFile)
	if err != nil {
		log.Fatal(err)
	}

	likeImg, _, err := image.Decode(likeFile)
	if err != nil {
		log.Fatal(err)
	}

	dislikeImg, _, err := image.Decode(dislikeFile)
	if err != nil {
		log.Fatal(err)
	}

	commonImgs := []image.Image{bgImg, likeImg, dislikeImg}

	//2.2 decode tech icon
	var likeImgs []image.Image
	for _, v := range likeFiles {
		img, _, err := image.Decode(v)
		if err != nil {
			log.Fatal(err)
		}
		likeImgs = append(likeImgs, img)
	}

	//dotnetImg, _, err := image.Decode(dotnetFile)
	var dislikeImgs []image.Image
	for _, v := range dislikeFiles {
		img, _, err := image.Decode(v)
		if err != nil {
			log.Fatal(err)
		}
		dislikeImgs = append(dislikeImgs, img)
	}

	return commonImgs, likeImgs, dislikeImgs
}

func composeImage(saved *OutImages, cImgs, lImgs, dlImgs []image.Image) {
	//cImgs[0] => bgImg
	//cImgs[1] => likeImg
	//cImgs[2] => dislileImg

	//bg
	bgRectangle := image.Rectangle{image.Point{0, 0}, cImgs[0].Bounds().Size()}
	rgba := image.NewRGBA(bgRectangle)
	draw.Draw(rgba, bgRectangle, cImgs[0], image.Point{0, 0}, draw.Src)

	//position of img on bgImg
	likeXY := image.Point{000, 0}
	likeRectangle := image.Rectangle{likeXY, likeXY.Add(cImgs[1].Bounds().Size())}
	draw.Draw(rgba, likeRectangle, cImgs[1], image.Point{0, 0}, draw.Over)

	dislikeXY := image.Point{0, 198}
	dislikeRectangle := image.Rectangle{dislikeXY, dislikeXY.Add(cImgs[2].Bounds().Size())}
	draw.Draw(rgba, dislikeRectangle, cImgs[2], image.Point{0, 0}, draw.Over)

	//loop
	x := 520
	size, margin := calcSize(len(lImgs), 0)
	for _, v := range lImgs {
		xy := image.Point{x, 0}
		if len(lImgs) > 4 {
			//resize
			vv := resize.Resize(size, size, v, resize.Lanczos3)
			rRectangle := image.Rectangle{xy, xy.Add(vv.Bounds().Size())}
			draw.Draw(rgba, rRectangle, vv, image.Point{0, 0}, draw.Over)
			x += margin
		} else {
			//default
			rRectangle := image.Rectangle{xy, xy.Add(v.Bounds().Size())}
			draw.Draw(rgba, rRectangle, v, image.Point{0, 0}, draw.Over)
			x += 230
		}
	}

	x = 550
	size, margin = calcSize(len(dlImgs), -30)
	for _, v := range dlImgs {
		xy := image.Point{x, 198}
		if len(dlImgs) > 4 {
			//resize
			vv := resize.Resize(size, size, v, resize.Lanczos3)
			rRectangle := image.Rectangle{xy, xy.Add(vv.Bounds().Size())}
			draw.Draw(rgba, rRectangle, vv, image.Point{0, 0}, draw.Over)
			x += margin
		} else {
			rRectangle := image.Rectangle{xy, xy.Add(v.Bounds().Size())}
			draw.Draw(rgba, rRectangle, v, image.Point{0, 0}, draw.Over)
			x += 230
		}
	}

	//savedFile := "./images/saved.png"
	out, err := os.Create(*saved.File)
	if err != nil {
		log.Fatal(err)
	}

	switch *saved.Format {
	case "jpg", "jpeg":
		//1.jpeg
		var opt jpeg.Options
		opt.Quality = 100
		jpeg.Encode(out, rgba, &opt)
	case "png":
		//2.png
		png.Encode(out, rgba)
	}
}

func calcSize(num int, adjustment float64) (uint, int) {
	//max size is 1064
	var baseSize float64 = 1064 + adjustment
	var wholeW float64 = baseSize / float64(num)
	w := wholeW / 115.0 * 100.0
	return uint(w), int(wholeW)
}
