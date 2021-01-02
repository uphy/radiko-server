package main

import (
	"flag"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/uphy/radiko-server/api"
	"github.com/uphy/radiko-server/library"
)

var (
	baseURL   string
	dataDir   string
	staticDir string
)

func main() {
	flag.StringVar(&baseURL, "base", "http://localhost:8080/", "")
	flag.StringVar(&dataDir, "data", "data", "")
	flag.StringVar(&staticDir, "static", "static", "")
	flag.Parse()

	l := library.New(dataDir)

	if err := l.Load(); err != nil {
		panic(err)
	}

	e := echo.New()
	a := api.New(l, baseURL)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.POST("/recordings/record", a.Record)
	e.GET("/recordings/", a.List)
	e.GET("/recordings/recording/:stationID/:start", a.Get)
	e.GET("/recordings/recording/:stationID/:start/playlist", a.M3U8)
	e.GET("/recordings/recording/:stationID/:start/:file", a.File)
	e.Static("/", staticDir)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
