package proxy

import (
	"fmt"
	"net/http"
	"net/url"
    "time"
    "io"
	"strings"
	"strconv"
	"log"

	"github.com/MohOdejimi/ProxyCache/internal/cache"

	"github.com/julienschmidt/httprouter"
)

var store = cache.NewStore()

func init() {
	store.StartCleanup(10 * time.Minute)
}

func singleJoiningSlash(a, b string) string {
	aSlash := len(a) > 0 && a[len(a)-1] == '/'
	bSlash := len(b) > 0 && b[0] == '/'
	switch {
	case aSlash && bSlash:
		return a + b[1:]
	case !aSlash && !bSlash:
		if b == "" {
			return a
		}
		return a + "/" + b
	}
	return a + b
}

func shouldCacheResponse(req *http.Request, response *http.Response) bool {
	if req.Method != http.MethodGet {
		return false
	}
	if response.StatusCode != http.StatusOK {
		return false
	}

	cacheControl := response.Header.Get("Cache-Control")
	if strings.Contains(cacheControl, "no-cache") ||
	strings.Contains(cacheControl, "must-revalidate") ||
	strings.Contains(cacheControl, "proxy-revalidate") {
		return false
	}

	if expires := response.Header.Get("Expires"); expires != "" {
    if expTime, err := time.Parse(time.RFC1123, expires); err == nil && time.Now().After(expTime) {
        return false
  	  }
	}

	if response.Header.Get("Pragma") == "no-cache" {
    	return false
	}

	if contentLength := response.Header.Get("Content-Length"); contentLength != "" {
    if size, err := strconv.Atoi(contentLength); err == nil && size > 10*1024*1024 { 
        return false
   	 }
	}

	reqCacheControl := req.Header.Get("Cache-Control")
	if strings.Contains(reqCacheControl, "no-cache") || strings.Contains(reqCacheControl, "no-store") {
    	return false
	}

	if req.Header.Get("Authorization") != "" {
  	  return false
	}

	if response.Header.Get("Set-Cookie") != "" {
		return false 
	}
	return true
}

func parseMaxAge(cacheControl string) (time.Duration, bool) {
    for _, directive := range strings.Split(cacheControl, ",") {
        directive = strings.TrimSpace(directive)
        if strings.HasPrefix(directive, "max-age=") {
            if val, err := strconv.Atoi(strings.TrimPrefix(directive, "max-age=")); err == nil {
                return time.Duration(val) * time.Second, true
            }
        }
    }
    return 0, false
}

func directResponseToClient(w http.ResponseWriter, responseData *cache.ResponseData, cacheStatus string) {
	for key, value := range responseData.Headers {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}

	w.Header().Set("X-Cache", cacheStatus)
	w.WriteHeader(responseData.StatusCode)

	log.Println(cacheStatus)

	fmt.Println("Sending response with status:", responseData.StatusCode, "body length:", len(responseData.Body))
	_, err := w.Write(responseData.Body)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func ClearCache() {
	store.Clear()
}

func ProxyHandler(origin *url.URL) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		targetUrl := *origin
		targetUrl.Path = singleJoiningSlash(targetUrl.Path, r.URL.Path)
		targetUrl.RawQuery = r.URL.RawQuery

		req, err := http.NewRequest(r.Method, targetUrl.String(), r.Body)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}
		req.Header = r.Header.Clone()

		cacheKey := cache.GenerateCacheKey(r.Method, targetUrl.String())
		cachedResponse := store.Get(cacheKey)
		
		if cachedResponse == nil {
			resp, err := ForwardToOriginServer(req)
			if err != nil {
				http.Error(w, "Failed to reach origin", http.StatusBadGateway)
				return
			}
            var defaultTTL time.Duration

			maxAge, hasMaxAge := parseMaxAge(resp.Header.Get("Cache-Control"))
			if hasMaxAge {
				defaultTTL = maxAge
			} else {
				defaultTTL = 5 * time.Minute
			}

            if shouldCacheResponse(req, resp) {
                responseData, err := store.Set(cacheKey, resp, defaultTTL)
                if err != nil {
                    http.Error(w, "Failed to cache response", http.StatusInternalServerError)
                    return
                }
                directResponseToClient(w, responseData, "MISS")

            } else {
                responseData := &cache.ResponseData{
                    StatusCode: resp.StatusCode,
                    Headers:    resp.Header.Clone(),
                }
                body, err := io.ReadAll(resp.Body)
                if err != nil {
                    http.Error(w, "Failed to read response body", http.StatusInternalServerError)
                    return
                }
                responseData.Body = body
                directResponseToClient(w, responseData, "MISS")
            }
		} else {
			directResponseToClient(w, cachedResponse, "HIT")
		}
	}
}
