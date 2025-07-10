package routes

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
)

var (
	once    sync.Once
	proxies map[string]*httputil.ReverseProxy
)

func initProxies() {
	proxies = make(map[string]*httputil.ReverseProxy)

	serviceEnv := map[string]string{
		"admin":     os.Getenv("ADMIN_SERVICE_URL"),
		"validator": os.Getenv("VALIDATOR_SERVICE_URL"),
		"web3":      os.Getenv("WEB3_VALIDATOR_URL"),
	}

	for prefix, raw := range serviceEnv {
		if raw == "" {
			continue
		}
		target, err := url.Parse(raw)
		if err != nil {
			panic(fmt.Sprintf("Invalid url for %s: %v", prefix, err))
		}
		proxies[prefix] = httputil.NewSingleHostReverseProxy(target)
	}
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	once.Do(initProxies)

	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)
	service := parts[0]
	proxy, ok := proxies[service]
	if !ok {
		http.Error(w, "Unknown service: "+service, http.StatusBadGateway)
		return
	}

	newPath := "/"
	if len(parts) == 2 {
		newPath += parts[1]
	}
	r.URL.Path = newPath

	proxy.ServeHTTP(w, r)
}
