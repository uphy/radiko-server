package api

import (
	"net/http"
	"os"

	"github.com/labstack/echo"
)

func (a *API) File(c echo.Context) error {
	stationID := c.Param("stationID")
	start := c.Param("start")
	file := c.Param("file")
	startTime, err := a.library.ParseTime(start)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid 'start'")
	}
	f := a.library.File(stationID, startTime, file)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return echo.NewHTTPError(http.StatusNotFound, "file not found")
	}
	return c.File(f)
}
