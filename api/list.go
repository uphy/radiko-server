package api

import (
	"net/http"

	"github.com/labstack/echo"
)

func (a *API) List(c echo.Context) error {
	recordings, err := a.library.List()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get the list of recording files")
	}
	return c.JSON(http.StatusOK, recordings)
}
