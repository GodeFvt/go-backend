package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
)

func SentryCapture(e *echo.Echo) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if hub := sentryecho.GetHubFromContext(c); hub != nil {
			if reflect.TypeOf(err).String() == "*echo.HTTPError" {
				if echoErr := err.(*echo.HTTPError); echoErr != nil {
					if echoErr.Code == http.StatusNotFound {
						hub.WithScope(func(scope *sentry.Scope) {
							scope.SetTag("method", string(c.Request().Method))
							scope.SetTag("path", c.Path())
							scope.SetTag("url", c.Request().URL.RawPath)
							scope.SetLevel(sentry.LevelWarning)
							hub.CaptureMessage(
								fmt.Sprintf(
									"user try to access with method: %s on path: %s",
									c.Request().Method,
									c.Path(),
								),
							)
						})
					}
					if echoErr.Code >= http.StatusInternalServerError {
						if reflect.TypeOf(echoErr.Message) != nil {
							if reflect.TypeOf(echoErr.Message).Kind() == reflect.String {
								hub.WithScope(func(scope *sentry.Scope) {
									bodyData := c.Get("params")
									if bodyData != nil {
										bu, _ := json.Marshal(bodyData)
										scope.SetRequestBody(bu)
									}

									scope.SetTag("method", string(c.Request().Method))
									scope.SetTag("path", c.Path())
									scope.SetTag("url", c.Request().URL.RawPath)
									scope.SetExtra("jwt-payload", c.Get("payload"))
									scope.SetLevel(sentry.LevelError)
									hub.CaptureMessage(
										fmt.Sprintf(
											"[ERROR %d] %s",
											echoErr.Code,
											echoErr.Message.(string),
										),
									)
								})
							} else if reflect.TypeOf(echoErr.Message).Kind() == reflect.Map {
								hub.WithScope(func(scope *sentry.Scope) {
									bodyData := c.Get("params")
									if bodyData != nil {
										bu, _ := json.Marshal(bodyData)
										scope.SetRequestBody(bu)
									}
									var msg string
									if m, ok := echoErr.Message.(map[string]interface{}); ok {
										if v, exists := m["msg"]; exists {
											msg = cast.ToString(v)
										}
										if v, exists := m["message"]; exists {
											msg = cast.ToString(v)
										}
									}

									scope.SetTag("method", string(c.Request().Method))
									scope.SetTag("path", c.Path())
									scope.SetTag("url", c.Request().URL.RawPath)
									scope.SetExtra("jwt-payload", c.Get("payload"))
									scope.SetLevel(sentry.LevelError)
									hub.CaptureMessage(fmt.Sprintf("[ERROR %d] %s", echoErr.Code, msg))
								})
							}
						}
					}
				}
			} else if err != nil {
				customCaptureException(c, hub, err, sentry.LevelFatal)
			}
		}
		e.DefaultHTTPErrorHandler(err, c)
	}
}

func customCaptureException(c echo.Context, hub *sentry.Hub, err error, level sentry.Level) {
	prefixMsg := "INFO"
	switch level {
	case sentry.LevelError:
		prefixMsg = "ERROR"
	case sentry.LevelFatal:
		prefixMsg = "PANIC"
	case sentry.LevelWarning:
		prefixMsg = "WARNING"
	case sentry.LevelDebug:
		prefixMsg = "DEBUG"
	}

	hub.WithScope(func(scope *sentry.Scope) {
		bodyData := c.Get("params")
		scope.SetExtra("jwt-payload", c.Get("payload"))
		if bodyData != nil {
			bu, _ := json.Marshal(bodyData)
			scope.SetRequestBody(bu)
		}

		event := &sentry.Event{
			Tags: map[string]string{
				"method": string(c.Request().Method),
				"path":   c.Path(),
				"url":    c.Request().URL.RawPath,
			},
			Level:   level,
			Message: err.Error(),
			Exception: []sentry.Exception{
				{
					Type:       fmt.Sprintf("[%s] %s", prefixMsg, err.Error()),
					Value:      reflect.TypeOf(err).String(),
					Stacktrace: sentry.NewStacktrace(),
				},
			},
		}

		event = scope.ApplyToEvent(event, &sentry.EventHint{OriginalException: err}, hub.Client())
		hub.CaptureEvent(event)
	})
}
