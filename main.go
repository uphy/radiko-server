package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
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
	migrate      bool
)

func main() {
	flag.StringVar(&baseURL, "base", "http://localhost:8080/", "")
	flag.StringVar(&dataDir, "data", "data", "")
	flag.StringVar(&staticDir, "static", "static", "")
	flag.StringVar(&relativePath, "rel", "", "")
	flag.IntVar(&port, "port", 8080, "")
	flag.BoolVar(&migrate, "migrate", false, "")
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

	if err := updateStaticBase(); err != nil {
		panic(err)
	}

	l, err := library.New(dataDir)
	if err != nil {
		panic(err)
	}

	if err := l.Load(); err != nil {
		panic(err)
	}

	if migrate {
		if err := l.Migrate(); err != nil {
			log.Errorf("Failed to migrate: %v", err)
		}
	}

	go func() {
		for {
			l.ScanAndRecord()
			time.Sleep(time.Hour)
		}
	}()

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
	e.GET(relativePath+"/recordings/recording/:stationID/:start/audio", a.Audio)
	e.GET(relativePath+"/recordings/recording/:stationID/:start/:file", a.File)
	if len(relativePath) == 0 {
		e.Static("/", staticDir)
	} else {
		e.Static(relativePath, staticDir)
	}

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

func updateStaticBase() error {
	indexFile := filepath.Join(staticDir, "index.html")
	b, err := ioutil.ReadFile(indexFile)
	if err != nil {
		return err
	}
	r := regexp.MustCompile(`<base href="(.*?)">`)
	rel := relativePath
	if !strings.HasSuffix(rel, "/") {
		rel = rel + "/"
	}
	replaced := r.ReplaceAll(b, []byte(fmt.Sprintf(`<base href="%s">`, rel)))
	return ioutil.WriteFile(indexFile, replaced, 0777)
}
