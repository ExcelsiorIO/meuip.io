package main

import (
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	addr := flag.String("addr", ":3000", "Listen address")
	flag.Parse()

	e := echo.New()
	e.Use(middleware.Logger())
	e.GET("/", ipHandler)
	e.GET("/debug", debugHandler)

	if err := e.Start(*addr); err != nil {
		panic(err)
	}
}

func ipHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, realIPFromRequest(ctx.Request()))
}

func debugHandler(ctx echo.Context) error {
	h := ctx.Request().Header
	return ctx.String(http.StatusOK, pretty(h))
}

func realIPFromRequest(r *http.Request) string {
	const (
		headerXForwardedFor = "X-Forwarded-For"
		headerXRealIP       = "X-Real-IP"
	)

	if ip := r.Header.Get(headerXForwardedFor); ip != "" {
		return strings.Split(ip, ", ")[0]
	}
	if ip := r.Header.Get(headerXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ra
}

func pretty(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
