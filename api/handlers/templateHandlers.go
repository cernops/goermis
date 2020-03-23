package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// TEMPLATE HANDLERS

//CreateHandler handles home page
func CreateHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "create.html", map[string]interface{}{
		"Auth": true,
		"csrf": c.Get("csrf"),
	})
}

//DeleteHandler handles home page
func DeleteHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "delete.html", map[string]interface{}{
		"Auth": true,
		"csrf": c.Get("csrf"),
	})
}

//DisplayHandler handles home page
func DisplayHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "display.html", map[string]interface{}{
		"Auth": true,
		"csrf": c.Get("csrf"),
	})
}

//HomeHandler handles home page
func HomeHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"Auth": true,
	})
}

//LogsHandler handles home page
func LogsHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "logs.html", map[string]interface{}{
		"Auth": true,
		"csrf": c.Get("csrf"),
	})
}

//ModifyHandler handles modify page
func ModifyHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "modify.html", map[string]interface{}{
		"Auth": true,
		"csrf": c.Get("csrf"),
	})

}
