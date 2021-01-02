package api

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo"
)

func (a *API) M3U8(c echo.Context) error {
	stationID := c.Param("stationID")
	start := c.Param("start")
	startTime, err := a.library.ParseTime(start)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid 'start'")
	}

	buf := new(bytes.Buffer)
	if err := a.library.GenerateM3U8(a.baseURL, stationID, startTime, buf); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate m3u8")
	}
	header := c.Response().Header()
	header.Add("Content-Type", "application/x-mpegURL")

	return c.String(200, buf.String())
}
