package proxy

import (
	"net/http"
	"time"
)

func ForwardToOriginServer(r *http.Request) (*http.Response, error) {
	var httpClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
