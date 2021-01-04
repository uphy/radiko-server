package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/uphy/radiko-server/api"
	"github.com/uphy/radiko-server/library"
)

var (
	// must ends with "/"
	baseURL   string
	dataDir   string
	staticDir string
	// empty, or must starts with "/", also must not ends with "/"
	// "": valid
	// "/": not valid
	// "/foo": valid
	// "/foo/": not valid
	// "foo": not valid
	relativePath string
	port         int
)

func main() {
	flag.StringVar(&baseURL, "base", "http://localhost:8080/", "")
	flag.StringVar(&dataDir, "data", "data", "")
	flag.StringVar(&staticDir, "static", "static", "")
	flag.StringVar(&relativePath, "rel", "", "")
	flag.IntVar(&port, "port", 8080, "")
	flag.Parse()

	if len(relativePath) != 0 {
		if !strings.HasPrefix(relativePath, "/") {
			relativePath = "/" + relativePath
		}
		if strings.HasSuffix(relativePath, "/") {
			relativePath = relativePath[0 : len(relativePath)-1]
		}
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL + "/"
	}

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
	e.POST(relativePath+"/recordings/record", a.Record)
	e.GET(relativePath+"/recordings/", a.List)
	e.GET(relativePath+"/recordings/recording/:stationID/:start", a.Get)
	e.GET(relativePath+"/recordings/recording/:stationID/:start/playlist", a.M3U8)
	e.GET(relativePath+"/recordings/recording/:stationID/:start/:file", a.File)
	if len(relativePath) == 0 {
		e.Static("/", staticDir)
	} else {
		e.Static(relativePath, staticDir)
	}

	//a.Static(e, relativePath, staticDir)

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}
