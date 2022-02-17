package gosession

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"gopkg.in/go-playground/validator.v9"
)

var e = echo.New()
var v = validator.New()

// Start - Starts the program
func Start() {

	// TODO can we make this private?
	e.Static("/", "public")
	startViperConfiguration()

	configureDatabases()

	// Configure Middlewares
	configureDefaultMiddlewares(e)

	// Configure Routes
	configureRoutes()

	e.Logger.Fatal(e.Start(fmt.Sprintf(config.AppEnv+":%s", config.Port)))
}

// configureDatabase will setup the mongo database connection
func configureDatabases() {
	connectToDatabase()
	connectToRedisLimiterDatabase()
	connectToRedisSessionDatabase()
}

// ConfigureRoutes will make calls to configure all the different routes for fiber
func configureRoutes() {

	configureDefaultRoutes()
	configureUserRoutes()
	configureAuthenticationRoutes()
	configureS3Routes()
}
