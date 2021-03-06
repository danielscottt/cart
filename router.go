package cart

import (
	"net/http"
	"regexp"
	"strings"
)

type route struct {
	path     string
	method   string
	callback RouterCallback
	params   map[string]string
	regex    *regexp.Regexp
	nodes    []string
	urlVars  []string
}

type router struct {
	rootHandlers    map[string]*route
	routes          *branch
	notFoundHandler RouterCallback
}

func (r *router) ServeHTTP(rsp http.ResponseWriter, req *http.Request) {

	if req.URL.Path == "/" {
		rt := r.rootHandlers[req.Method]
		rt.callback(req, rsp, rt.params)
		return
	} else {
		branch := r.routes.Find(strings.Split(string(req.URL.Path[1:]), "/"), req.Method)

		if branch != nil {
			rt := branch.key
			match := rt.regex.FindAllStringSubmatch(req.URL.Path, -1)
			if len(match) > 0 {
				for i, m := range match[0][1:] {
					rt.params[rt.urlVars[i]] = m
				}
				rt.callback(req, rsp, rt.params)
				return
			}
		}
	}

	rsp.WriteHeader(http.StatusNotFound)
	r.notFoundHandler(req, rsp, map[string]string{})
}

func (r *router) AddToRoutes(path string, callback RouterCallback, method string) {
	rt := &route{
		path:     path,
		method:   method,
		callback: callback,
		params:   make(map[string]string),
		urlVars:  make([]string, 0),
	}
	if path == "/" {
		r.rootHandlers[method] = rt
	} else {
		rt.nodes = strings.Split(string(path[1:]), "/")
		for i, n := range rt.nodes {
			if string(n[0]) == ":" {
				rt.urlVars = append(rt.urlVars, string(n[1:]))
				rt.nodes[i] = "([a-zA-Z0-9]+)"
			}
		}
		rt.regex = regexp.MustCompile("^/" + strings.Join(rt.nodes, "/") + "$")
		r.routes.Insert(rt)
	}
}
