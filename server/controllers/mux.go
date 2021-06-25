package controllers

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/VanjaRo/reminder-cli/server/models"
	"github.com/VanjaRo/reminder-cli/server/transport"
)

const paramsKey = "ps"

type ctxKey string

type urlParam struct {
	name     string
	regEx    string
	value    string
	position int
}

// rote stands for Handler route
type route struct {
	path    string
	method  string
	params  map[string]urlParam
	handler http.Handler
}

func (r *route) populate(req *http.Request) string {
	urlSlice := splitUrl(req.URL.Path)
	pathSlice := splitUrl(r.path)
	if len(urlSlice) != len(pathSlice) {
		return ""
	}
	for name, param := range r.params {
		regExParamVal := urlSlice[param.position]
		regEx := regexp.MustCompile(param.regEx)
		if name != "" && regEx.MatchString(regExParamVal) {
			param.value = regExParamVal
			r.params[name] = param
			pathSlice[param.position] = regExParamVal
		}
	}
	pathStr := "/" + strings.Join(pathSlice, "/")
	if req.URL.Path == pathStr {
		return r.method + pathStr
	}
	return ""
}

type RegExMux struct {
	routes    []*route
	routesMap map[string]*route
}

// parses URL parameters given a {param}:regex expression
func (h RegExMux) parseParam(url, regExParam string) urlParam {
	r := regexp.MustCompile(`({[a-z]+}:)(.+)`)
	matches := r.FindStringSubmatch(regExParam)
	// 1 – entire match
	// 2 – 1st group -> param name
	// 3 – 2nd group -> pram regex
	if len(matches) < 3 {
		return urlParam{
			regEx: ".+",
		}
	}
	replacer := strings.NewReplacer(
		"{", "",
		"}", "",
		":", "",
	)
	name := replacer.Replace(matches[1])
	regEx := matches[2]
	var position int
	for i, v := range splitUrl(url) {
		if v == matches[1]+matches[2] {
			position = i
		}
	}
	return urlParam{
		name:     name,
		regEx:    regEx,
		position: position,
	}
}

func (h RegExMux) params(url string) map[string]urlParam {
	ps := map[string]urlParam{}
	for _, v := range splitUrl(url) {
		p := h.parseParam(url, v)
		if p.name != "" {
			ps[p.name] = p
		}
	}
	return ps
}

func (h *RegExMux) Handle(method, pattern string, handler http.Handler) {
	ps := h.params(pattern)
	r := &route{
		method:  method,
		path:    pattern,
		params:  ps,
		handler: handler,
	}
	h.routes = append(h.routes, r)
}

func (h *RegExMux) Get(pattern string, handler http.Handler) {
	h.Handle(http.MethodGet, pattern, handler)
}

func (h *RegExMux) Post(pattern string, handler http.Handler) {
	h.Handle(http.MethodPost, pattern, handler)
}

func (h *RegExMux) Patch(pattern string, handler http.Handler) {
	h.Handle(http.MethodPatch, pattern, handler)
}

func (h *RegExMux) Put(pattern string, handler http.Handler) {
	h.Handle(http.MethodPut, pattern, handler)
}

func (h *RegExMux) Delete(pattern string, handler http.Handler) {
	h.Handle(http.MethodDelete, pattern, handler)
}

func (h RegExMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.routesMap = map[string]*route{}
	for _, route := range h.routes {
		key := route.populate(r)
		h.routesMap[key] = route
	}
	key := r.Method + r.URL.Path
	route, ok := h.routesMap[key]
	if !ok {
		transport.SendError(w, models.NotFoundError{})
		return
	}
	ctx := r.Context()
	if len(route.params) != 0 {
		ctx = context.WithValue(ctx, ctxKey(paramsKey), route.params)
	}
	route.handler.ServeHTTP(w, r.WithContext(ctx))
}

func splitUrl(s string) []string {
	var ret []string
	for _, p := range strings.Split(strings.TrimSpace(s), "/") {
		if strings.TrimSpace(p) != "" {
			ret = append(ret, p)
		}
	}
	return ret
}
