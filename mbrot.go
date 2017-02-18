package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math/cmplx"
	"math/rand"
	"net/http"
	"text/template"
	"strconv"
)

const (
	ViewWidth = 640
	ViewHeight = 480
	MaxEscape = 64
)

var indexTemplate *template.Template
var palette []color.RGBA
var escapeColor color.RGBA


func init() {
	var err error
	indexTemplate, err = template.ParseFiles("index.html")
	if err != nil {
		panic(err)
	}

	palette = make([]color.RGBA, MaxEscape)
	for i := 0; i < MaxEscape - 1; i++ {
		palette[i] = color.RGBA{uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256)), 255}
	}
	escapeColor = color.RGBA{0, 0, 0, 0}
}


func escape(c complex128) int {
	z := c
	for i := 0; i < MaxEscape - 1; i++ {
		if cmplx.Abs(z) > 2 {
			return i
		}
		z = cmplx.Pow(z, 2) + c
	}
	return MaxEscape - 1
}


func pixelToPoint(viewCenter complex128, x, y, imgWidth, imgHeight int, zoomWidth float64) complex128 {
	pixelWidth := zoomWidth / float64(imgWidth)
	pixelHeight := pixelWidth
	viewHeight := (float64(imgHeight) / float64(imgWidth)) * zoomWidth
	left := (real(viewCenter) - (zoomWidth / 2)) + pixelWidth / 2
	top := (imag(viewCenter) - (viewHeight / 2)) + pixelHeight / 2
	return complex(left + float64(x)*pixelWidth, top + float64(y)*pixelHeight)
}


func generate(imgWidth int, imgHeight int, viewCenter complex128, radius float64) image.Image {
	m := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	zoomWidth := radius * 2
	pixelWidth := zoomWidth / float64(imgWidth)
	pixelHeight := pixelWidth
	viewHeight := (float64(imgHeight) / float64(imgWidth)) * zoomWidth
	left := (real(viewCenter) - (zoomWidth / 2)) + pixelWidth / 2
	top := (imag(viewCenter) - (viewHeight / 2)) + pixelHeight / 2
	for x := 0; x < imgWidth; x++ {
		for y := 0; y < imgHeight; y++ {
			coord := complex(left + float64(x)*pixelWidth, top + float64(y)*pixelHeight)
			f := escape(coord)
			if f == MaxEscape -1 {
				m.Set(x, y, escapeColor)
			}
			m.Set(x, y, palette[f])
		}
	}
	return m
}


func SafeInt(s string, min, max, def int) int {
	i, err := strconv.Atoi(s)
    if err != nil {
		return def
    }
	if i < min || i >= max {
		return def
	}
	return i
}


func SafeFloat64(s string, def float64) float64 {
	f, err := strconv.ParseFloat(s, 64)
    if err != nil {
		return def
    }
	return f
}


func pic(w http.ResponseWriter, r *http.Request) {
	mx := SafeFloat64(r.FormValue("mx"), 0.0)
	my := SafeFloat64(r.FormValue("my"), 0.0)
	radius := SafeFloat64(r.FormValue("radius"), 2.0)

	m := generate(ViewWidth, ViewHeight, complex(mx, my), radius)
	w.Header().Set("Content-Type", "image/png")
	err := png.Encode(w, m)
	if err != nil {
		log.Println("png.Encode:", err)
	}
}


func zoom(w http.ResponseWriter, r *http.Request) {
	m := generate(ViewWidth, ViewHeight, complex128(0), 4)
	w.Header().Set("Content-Type", "image/png")
	err := png.Encode(w, m)
	if err != nil {
		log.Println("png.Encode:", err)
	}
}


func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	vars := make(map[string]interface{})
	vars["Title"] = "Mandelbrot"
	vars["Height"] = ViewHeight
	vars["Width"] = ViewWidth
	indexTemplate.Execute(w, vars)
}


func main() {
	log.Println("Listening")
	defer log.Println("Exiting")

	http.HandleFunc("/", index)
	http.HandleFunc("/pic", pic)

	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}