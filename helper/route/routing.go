package route

import (
	"github.com/labstack/echo/v4"
)

func RegisterVersion(e *echo.Echo) {
	e.GET("/version", Version)
}
