package router

import (
	"github.com/labstack/echo/v4"
	"gitlab.cern.ch/lb-experts/goermis/api/handlers"
)

//InitRoutes initializes the routes
func InitRoutes(e *echo.Echo) {

	e.GET("/", handlers.HomeHandler)
	e.GET("/lbweb/", handlers.HomeHandler)
	e.GET("/lbweb/create/", handlers.CreateHandler)
	e.GET("/lbweb/modify/", handlers.ModifyHandler)
	e.GET("/lbweb/display/", handlers.DisplayHandler)
	e.GET("/lbweb/delete/", handlers.DeleteHandler)
	e.GET("/lbweb/logs/", handlers.LogsHandler)

	e.GET("/lbweb/aliases", handlers.GetAliases)
	e.GET("/lbweb/aliases/:alias", handlers.GetAlias)

	e.GET("/lbweb/*/checkname/:hostname", handlers.CheckNameDNS)
	e.POST("/lbweb/new_alias", handlers.NewAlias)
	e.POST("/lbweb/delete_alias", handlers.DeleteAlias)
	e.POST("/lbweb/modify_alias", handlers.ModifyAlias)
}
