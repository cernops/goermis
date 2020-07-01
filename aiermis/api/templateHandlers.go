package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// TEMPLATE HANDLERS

//CreateHandler handles home page
func CreateHandler(c echo.Context) error {
	return defaultHandler(c, "create.html")
}

//DeleteHandler handles home page
func DeleteHandler(c echo.Context) error {
	return defaultHandler(c, "delete.html")
}

//DisplayHandler handles home page
func DisplayHandler(c echo.Context) error {
	return defaultHandler(c, "display.html")
}

//HomeHandler handles home page
func HomeHandler(c echo.Context) error {
	return defaultHandler(c, "home.html")
}

//LogsHandler handles home page
func LogsHandler(c echo.Context) error {
	return defaultHandler(c, "logs.html")
}

//ModifyHandler handles modify page
func ModifyHandler(c echo.Context) error {
	return defaultHandler(c, "modify.html")

}

func defaultHandler(c echo.Context, page string) error {
	return c.Render(http.StatusOK, page, map[string]interface{}{
		"Auth": true,
		"csrf": c.Get("csrf"),
		"User": c.Request().Header.Get("X-Forwarded-User"),
	})
}
