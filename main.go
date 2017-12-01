package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

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

	m.Group("/r/(?P<name>[a-zA-Z]{3,})", func(r martini.Router) {
		m.Any("", ProxyToRegisteredIP)
		m.Any("/**", ProxyToRegisteredIP)
	})

	m.Group("/r/(?P<name>[a-zA-Z]{3,}):(?P<port>[0-9]+)", func(r martini.Router) {
		m.Any("", ProxyToRegisteredIP)
		m.Any("/**", ProxyToRegisteredIP)
	})

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

func ProxyToRegisteredIP(r *http.Request, p martini.Params) (int, string) {
	name := p["name"]
	registeredIP, exists := registered[name]

	if !exists {
		return http.StatusNotFound, "Not found. " + name + " is not currently registered."
	}

	port := p["port"]
	if port == "" {
		port = "80"
	}
	restoUrl := p["_1"]

	tr := &http.Transport{}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 3,
	}

	req, err := http.NewRequest(
		r.Method,
		fmt.Sprintf(
			"http://%s:%s/%s",
			registeredIP.LastIP,
			port,
			restoUrl,
		),
		r.Body,
	)

	req.URL.RawQuery = r.URL.RawQuery

	req.ContentLength = r.ContentLength
	req.Form = r.Form
	req.Header = r.Header
	req.MultipartForm = r.MultipartForm
	req.PostForm = r.PostForm
	req.Proto = r.Proto
	req.ProtoMajor = r.ProtoMajor
	req.ProtoMinor = r.ProtoMinor
	req.TLS = r.TLS
	req.Trailer = r.Trailer
	req.TransferEncoding = r.TransferEncoding

	if err != nil {
		return http.StatusInternalServerError, "Internal Server Error"
	}

	resp, err := client.Do(req)
	if err != nil {
		return http.StatusInternalServerError, "Internal Server Error. Hmm... Host is offline?"
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return http.StatusInternalServerError, "Internal Server Error"
	}

	return resp.StatusCode, string(b)
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

func GetIPFromRequest(r *http.Request) string {
	ra := r.RemoteAddr
	if ip := r.Header.Get(HeaderXForwardedFor); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := r.Header.Get(HeaderXRealIP); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}
