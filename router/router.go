package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gitlab.cern.ch/lb-experts/goermis/api/handlers"
	m "gitlab.cern.ch/lb-experts/goermis/api/middleware"
)

//New Echo COntext
func New() *echo.Echo {

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	lbweb := e.Group("/lbweb")
	lbweb.Use(m.CheckAuthorization)
	lbweb.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		Skipper: middleware.DefaultSkipper, TokenLength: 32,
		TokenLookup: "form:csrf", ContextKey: "csrf", CookieName: "_csrf", CookieMaxAge: 86400,
	}))

	lbweb.GET("/", handlers.HomeHandler)
	lbweb.GET("/create/", handlers.CreateHandler)
	lbweb.GET("/modify/", handlers.ModifyHandler)
	lbweb.GET("/display/", handlers.DisplayHandler)
	lbweb.GET("/delete/", handlers.DeleteHandler)
	lbweb.GET("/logs/", handlers.LogsHandler)
	lbweb.POST("/new_alias", handlers.NewAlias)
	lbweb.POST("/delete_alias", handlers.DeleteAlias)
	lbweb.POST("/modify_alias", handlers.ModifyAlias)
	lbweb.GET("/*/checkname/:hostname", handlers.CheckNameDNS)

	api := e.Group("/api/v1")
	api.Use(m.CheckAuthorization)
	api.GET("/aliases", handlers.GetAliases)
	api.GET("/aliases/:alias", handlers.GetAlias)
	api.DELETE("/aliases", handlers.DeleteAlias)
	api.POST("/aliases", handlers.NewAlias)
	return e
}
