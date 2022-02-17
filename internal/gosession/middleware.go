package gosession

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ulule/limiter/v3"
)

// ConfigureMiddlewares will make calls to configure all the different middlewares
func configureDefaultMiddlewares(e *echo.Echo) {
	// This runs before the router
	e.Pre(middleware.RemoveTrailingSlash())
	// Use this ID to track the route through the microservices for logging, etc
	e.Pre(middleware.RequestID())
	//e.Use(middleware.CSRF())
	// only enable this on certain route groups

	// TODO use gorillas implentation instead
	// e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
	// 	TokenLookup:  "header:X-XSRF-TOKEN",
	// 	CookieSecure: true,
	// }))

	// TODO
	//e.Use(middleware.CORS())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:4200", "http://127.0.0.1:4200"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAccessControlAllowCredentials, echo.HeaderCookie, echo.HeaderSetCookie},
		ExposeHeaders:    []string{echo.HeaderSetCookie},
		AllowCredentials: true,
	}))
	//e.Use(middleware.Gzip())
	//e.Use((middleware.HTTPSRedirect()))
	//e.Use((middleware.Secure())) TODO readd
	// TODO change the secret to be env passed in via kubernetes secrets
	//e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	// TODO setup the auth model and the auth policy
	// check the role from the cookie
	// TODO is checking the role from the cookie secure?
	// If not should we go to one where the session token is secure?
	// OR else also store in the redis store the sessionID, the userSessionID and also the role
	// enforcer, err := casbin.NewEnforcer("casbin_auth_model.conf", "casbin_auth_policy.csv")
	// if err != nil {
	// 	fmt.Println("Error")
	// }
	// e.Use(casbin_mw.Middleware(enforcer))
}

// Custom Middlewares -----------------------------------------------------------------------

var (
	ipRateLimiter *limiter.Limiter
	store         limiter.Store
)

// IPRateLimit - This will limit the amount of times a specific user (IP) can add an endpoint
// customize so you can pass in the variables to define custom limits per endpoint
// for limiting user not ip: https://github.com/ulule/limiter/issues/44
// if break rule too many times need to start increasing up to a max limit
// ask on stack and show use cases
// https://auth0.com/docs/policies/rate-limit-policy/database-connections-rate-limits
// https://timoh6.github.io/2015/05/07/Rate-limiting-web-application-login-attempts.html
func IPRateLimit(limit int64, period time.Duration) echo.MiddlewareFunc {
	// 1. Configure
	rate := limiter.Rate{
		Period: period,
		Limit:  limit,
	}

	store := redisLimiterInstance.Store

	ipRateLimiter = limiter.New(store, rate)

	// 2. Return middleware handler
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			ip := c.RealIP()
			limiterCtx, err := ipRateLimiter.Get(c.Request().Context(), ip)
			if err != nil {
				log.Printf("IPRateLimit - ipRateLimiter.Get - err: %v, %s on %s", err, ip, c.Request().URL)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"success": false,
					"message": err,
				})
			}

			h := c.Response().Header()

			h.Set("X-RateLimit-Limit", strconv.FormatInt(limiterCtx.Limit, 10))
			h.Set("X-RateLimit-Remaining", strconv.FormatInt(limiterCtx.Remaining, 10))
			h.Set("X-RateLimit-Reset", strconv.FormatInt(limiterCtx.Reset, 10))

			if limiterCtx.Reached {
				log.Printf("Too Many Requests from %s on %s", ip, c.Request().URL)
				return c.JSON(http.StatusTooManyRequests, echo.Map{
					"success": false,
					"message": "Too many calls to this endpoint, please try again later",
				})
			}

			return next(c)
		}
	}
}

// SessionMiddleware to confirm the user has a valid session
func SessionMiddleware(role string) echo.MiddlewareFunc {

	// 2. Return middleware handler
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {

			store := redisSessionInstance.Store

			session, err := store.Get(c.Request(), "session_")
			if err != nil {
				return c.JSON(http.StatusForbidden, "access denied")
			}

			if session.Values["userID"] == nil || session.Values["userID"] == "" {
				return c.JSON(http.StatusForbidden, "access denied 1")
			}

			// pass in min role to use this route here.
			userRole := session.Values["role"]

			fmt.Println("Role passed into middleware: " + role)
			fmt.Println("UserRole determined: ", userRole)
			fmt.Println("-------------------------------------")

			if role == "admin" {
				// only allow users with role as admin to access this route
				if userRole != "admin" {
					return c.JSON(http.StatusForbidden, "access denied")
				}
			} else if role == "user" {
				// only allow users with role as user or admin to access this route
				if userRole != "admin" && userRole != "user" {
					return c.JSON(http.StatusForbidden, "access denied")
				}
			}

			return next(c)
		}
	}
}
