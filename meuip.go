package main

import (
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
	m.Get("/:name", GetRegisteredIP)
	m.Post("/:name/:key", RegisterMyIP)

	m.Run()
}

func GetMyIP(r *http.Request) string {
	return GetIPFromRequest(r)
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
		reg.LastIP = GetIPFromRequest(r)
	} else {
		registered[name] = NewRegister(name, key, GetIPFromRequest(r))
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

func GetIPFromRequest(r *http.Request) string {
	hostPort := strings.Split(r.RemoteAddr, ":")
	return hostPort[0]
}
