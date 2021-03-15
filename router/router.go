package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gitlab.cern.ch/lb-experts/goermis/ermis"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
)

var (
	log = bootstrap.GetLog()
)

//New Echo Context
func New() *echo.Echo {
	e := echo.New()
	//CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	//Recover
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{}))

	//UI routes
	lbweb := e.Group("/lbweb")

	//Custom middleware in API
	lbweb.Use(ermis.CheckAuthorization)
	//CSRF
	lbweb.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		Skipper: middleware.DefaultSkipper, TokenLength: 32,
		TokenLookup: "form:csrf", ContextKey: "csrf", CookieName: "_csrf", CookieMaxAge: 86400,
	}))

	lbweb.GET("/", ermis.HomeHandler)
	lbweb.GET("/create", ermis.CreateHandler)
	lbweb.GET("/modify", ermis.ModifyHandler)
	lbweb.GET("/display", ermis.DisplayHandler)
	lbweb.GET("/delete", ermis.DeleteHandler)
	lbweb.GET("/logs", ermis.LogsHandler)
	lbweb.POST("/new_alias", ermis.CreateAlias)
	lbweb.POST("/delete_alias", ermis.DeleteAlias)
	lbweb.POST("/modify_alias", ermis.ModifyAlias)
	lbweb.GET("/checkname", ermis.CheckNameDNS)

	//CLI routes
	entrypoint := e.Group("/p/api/v1")
	entrypoint.Use(ermis.CheckAuthorization)
	entrypoint.GET("/raw/", ermis.GetAliasRaw)
	entrypoint.GET("/alias/", ermis.GetAlias)
	entrypoint.DELETE("/alias/", ermis.DeleteAlias)
	entrypoint.POST("/alias/", ermis.CreateAlias)
	entrypoint.PATCH("/alias/:id/", ermis.ModifyAlias)

	//lbclients
	lbclient := e.Group("/lb/api/v1")
	lbclient.POST("/lbclient/", lbclient.CreateHost)
	lbclient.GET("/lbclient/", lbclient.GetHosts)

	return e
}
