package router

import (
	//"gitlab.cern.ch/lb-experts/goermis/api"
	//"gitlab.cern.ch/lb-experts/goermis/api/middlewares"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//New Echo COntext
func New() *echo.Echo {

	e := echo.New()
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://localhost:8080", "*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	//adminGroup := e.Group("/admin")
	//jwtGroup := e.Group("/jwt")

	/*middlewares.SetMainMiddlewares(e)
	middlewares.SetCompleteLogMiddlware(e)

	middlewares.SetAdminMiddlewares(adminGroup)
	middlewares.SetJwtMiddlewares(jwtGroup)

	api.MainGroup(e)

	api.AdminGroup(adminGroup)
	api.JwtGroup(jwtGroup)*/

	return e
}
