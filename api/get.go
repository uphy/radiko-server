package api

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/uphy/radiko-server/library"
)

type GetResponse struct {
	Status    *library.Status          `json:"status"`
	Recording *library.RecordingDetail `json:"recording"`
}

func (a *API) Get(c echo.Context) error {
	stationID := c.Param("stationID")
	start := c.Param("start")
	startTime, err := a.library.ParseTime(start)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid 'start'")
	}
	recording, err := a.library.Get(stationID, startTime)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get recording file")
	}
	status, err := a.library.GetStatus(stationID, startTime)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get recording file status")
	}
	return c.JSON(http.StatusOK, GetResponse{
		Status:    status,
		Recording: recording,
	})
}
