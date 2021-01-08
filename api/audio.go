package api

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo"
)

func (a *API) Audio(c echo.Context) error {
	stationID := c.Param("stationID")
	start := c.Param("start")
	format := c.QueryParam("format")
	startTime, err := a.library.ParseTime(start)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid 'start'")
	}

	if len(format) == 0 {
		format = "m3u8"
	}
	switch format {
	case "m3u8":
		buf := new(bytes.Buffer)
		if err := a.library.GenerateM3U8(a.baseURL, stationID, startTime, buf); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate m3u8")
		}
		header := c.Response().Header()
		header.Add("Content-Type", "application/x-mpegURL")
		return c.String(200, buf.String())
	case "mp3":
		return c.File(a.library.MP3(stationID, startTime))
	case "aac":
		return c.File(a.library.AAC(stationID, startTime))
	}
	return echo.NewHTTPError(http.StatusBadRequest, "No such format. Supported format are m3u8/mp3/aac: format="+format)
}
