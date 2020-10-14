package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"gitlab.cern.ch/lb-experts/goermis/aiermis/api"
	m "gitlab.cern.ch/lb-experts/goermis/aiermis/middleware"
)

//New Echo Context
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

	lbweb.GET("/", api.HomeHandler)
	lbweb.GET("/create", api.CreateHandler)
	lbweb.GET("/modify", api.ModifyHandler)
	lbweb.GET("/display", api.DisplayHandler)
	lbweb.GET("/delete", api.DeleteHandler)
	lbweb.GET("/logs", api.LogsHandler)
	lbweb.POST("/new_alias", api.CreateAlias)
	lbweb.POST("/delete_alias", api.DeleteAlias)
	lbweb.POST("/modify_alias", api.ModifyAlias)
	lbweb.GET("/checkname", api.CheckNameDNS)

	lbterm := e.Group("/p/api/v1")
	lbterm.Use(m.CheckAuthorization)
	lbterm.GET("/alias/", api.GetAlias)
	lbterm.DELETE("/alias/", api.DeleteAlias)
	lbterm.POST("/alias/", api.CreateAlias)
	lbterm.PATCH("/alias/:id/", api.ModifyAlias)

	return e
}
