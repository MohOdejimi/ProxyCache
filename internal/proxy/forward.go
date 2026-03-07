package proxy

import (
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func ForwardToOriginServer(r *http.Request) (*http.Response, error) {
	resp, err := httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
