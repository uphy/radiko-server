package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
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
	if err := a.library.Record(req.StationID, start); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to record: %s", err.Error()))
	}
	return nil
}
