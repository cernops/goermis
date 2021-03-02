package api

/*This file contains the template handlers*/
import (
	"github.com/labstack/echo/v4"
)

// TEMPLATE HANDLERS

//CreateHandler handles home page
func CreateHandler(c echo.Context) error {
	return MessageToUser(c, 200, "", "create.html")
}

//DeleteHandler handles home page
func DeleteHandler(c echo.Context) error {
	return MessageToUser(c, 200, "", "delete.html")
}

//DisplayHandler handles home page
func DisplayHandler(c echo.Context) error {
	return MessageToUser(c, 200, "", "display.html")
}

//HomeHandler handles home page
func HomeHandler(c echo.Context) error {
	return MessageToUser(c, 200, "", "home.html")
}

//LogsHandler handles home page
func LogsHandler(c echo.Context) error {
	return MessageToUser(c, 200, "", "logs.html")
}

//ModifyHandler handles modify page
func ModifyHandler(c echo.Context) error {
	return MessageToUser(c, 200, "", "modify.html")

}
