package route

import (
	"net/http"

	"github.com/GodeFvt/go-backend/helper"
	"github.com/labstack/echo/v4"
)

func HelloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func Version(c echo.Context) error {
	var goVersion = helper.GetENV("GOLANG_VERSION", "not found env on key 'GOLANG_VERSION'")
	var env_build_number = helper.GetENV("env_build_number", "not found env on key 'env_build_number'")
	var env_version = helper.GetENV("env_version", "not found env on key 'env_version'")
	responseData := map[string]interface{}{
		"go_version":   goVersion,
		"build_number": env_build_number,
		"version":      env_version,
	}
	return c.JSON(http.StatusOK, responseData)
}
