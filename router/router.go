package router

import (
	//"gitlab.cern.ch/lb-experts/goermis/api"
	//"gitlab.cern.ch/lb-experts/goermis/api/middlewares"
	validator "github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//CustomValidator declares the data validator
type CustomValidator struct {
	validator *validator.Validate
}

//New Echo COntext
func New() *echo.Echo {

	e := echo.New()
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(middleware.Logger())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		Skipper: middleware.DefaultSkipper, TokenLength: 32,
		TokenLookup: "form:csrf", ContextKey: "csrf", CookieName: "_csrf", CookieMaxAge: 86400,
	}))

	//e.Validator = &CustomValidator{validator: validator.New()}
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
