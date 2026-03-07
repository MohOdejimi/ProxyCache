package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
    "time"
    "io"

	"github.com/MohOdejimi/ProxyCache/internal/cache"

	"github.com/julienschmidt/httprouter"
)

var store = cache.NewStore()

func shouldCacheResponse(req *http.Request, response *http.Response) bool {
	if req.Method != http.MethodGet {
		return false
	}
	if response.StatusCode != http.StatusOK {
		return false
	}

	return true
}

func directResponseToClient(w http.ResponseWriter, responseData *cache.ResponseData, cacheStatus string) {
	for key, value := range responseData.Headers {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}

	w.Header().Set("X-Cache", cacheStatus)
	w.WriteHeader(responseData.StatusCode)

	fmt.Println("Sending response with status:", responseData.StatusCode, "body length:", len(responseData.Body))
	_, err := w.Write(responseData.Body)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func ProxyHandler(origin *url.URL) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		targetUrl := *origin
		targetUrl.Path = path.Join(targetUrl.Path, r.URL.Path)
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
            const defaultTTL = 5 * time.Minute

            if shouldCacheResponse(req, resp) {
                responseData, err := store.Set(cacheKey, resp, defaultTTL)
                if err != nil {
                    http.Error(w, "Failed to cache response", http.StatusInternalServerError)
                    return
                }
                directResponseToClient(w, responseData, "MISS")
            } else {
                // can't cache, just forward response to client
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
			fmt.Println("Cache Hit")
			directResponseToClient(w, cachedResponse, "HIT")
		}
	}
}
