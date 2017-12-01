package main

import (
	"net"
	"net/http"
	"strings"

	"github.com/go-martini/martini"
)

type Register struct {
	Name   string
	Key    string
	LastIP string
}

var registered = make(map[string]*Register, 0)

func main() {
	m := martini.Classic()

	m.Get("/", GetMyIP)
	m.Group("/(?P<name>[a-zA-Z]{3,})", func(r martini.Router) {
		m.Get("", GetRegisteredIP)
		m.Post("/:key", RegisterMyIP)
		m.Delete("/:key", DeleteMyIP)
	})

	m.Run()
}

func GetMyIP(r *http.Request) string {
	return RealIPFromRequest(r)
}

func GetRegisteredIP(r *http.Request, p martini.Params) (int, string) {
	name := p["name"]

	if r, exists := registered[name]; exists {
		return http.StatusOK, r.LastIP
	}

	return http.StatusNotFound, "Not found"
}

func RegisterMyIP(r *http.Request, p martini.Params) (int, string) {
	name := p["name"]
	key := p["key"]

	reg, exists := registered[name]
	if exists && reg.Key != key {
		return http.StatusBadRequest, "Bad key"
	}

	if exists {
		reg.LastIP = RealIPFromRequest(r)
	} else {
		registered[name] = NewRegister(name, key, RealIPFromRequest(r))
	}

	return http.StatusOK, "Success!"
}

func DeleteMyIP(r *http.Request, p martini.Params) (int, string) {
	name := p["name"]
	key := p["key"]

	reg, exists := registered[name]
	if exists && reg.Key != key {
		return http.StatusBadRequest, "Bad key"
	}

	if exists {
		delete(registered, name)
	}

	return http.StatusOK, "Success!"
}

func NewRegister(name, key, lastip string) *Register {
	return &Register{
		Name:   name,
		Key:    key,
		LastIP: lastip,
	}
}

const HeaderXForwardedFor = "X-Forwarded-For"
const HeaderXRealIP = "X-Real-IP"

func RealIPFromRequest(r *http.Request) string {
	if ip := r.Header.Get(HeaderXForwardedFor); ip != "" {
		return strings.Split(ip, ", ")[0]
	}
	if ip := r.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ra
}
