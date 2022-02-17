package gosession

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ConfigureDefaultRoutes - Configure all the default routes here
func configureDefaultRoutes() {

	e.GET("/", defaultRoute, SessionMiddleware("user"))

}

func defaultRoute(c echo.Context) error {

	return c.JSON(http.StatusOK, "VALID")
}
