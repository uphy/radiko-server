package api

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

type RecordRequest struct {
	StationID string `json:"stationId"`
	Start     string `json:"start"`
}

func (a *API) Record(c echo.Context) error {
	var req RecordRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	start, err := a.library.ParseTime(req.Start)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid 'start'")
	}
	go func() {
		if err := a.library.Record(req.StationID, start); err != nil {
			log.Error("Failed to record %s", err.Error())
		}
	}()
	return c.NoContent(http.StatusAccepted)
}
