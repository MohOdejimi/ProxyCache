package cache

import (
	"io"
	"net/http"
	"sync"
	"time"
	"fmt"
)

type ResponseData struct {
	StatusCode int
	Headers   http.Header
	Body []byte
	ExpiresAt time.Time
}

type Store struct {
	mu    sync.RWMutex
	data map[string]*ResponseData
}

// use this constructor when the - cache is cleared or when the application starts up

func NewStore()	*Store {
	return &Store{
		data: make(map[string]*ResponseData),
	}
}

func (s *Store) Get(key string) *ResponseData {
    s.mu.RLock()
    defer s.mu.RUnlock()

    entry, exists := s.data[key]
    if !exists {
        return nil
    }
    if time.Now().After(entry.ExpiresAt) {
        return nil 
    }
    return entry
}

func (s *Store) Set(key string, response *http.Response, ttl time.Duration) (*ResponseData, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    body, err := io.ReadAll(response.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }
    defer response.Body.Close()

    data := &ResponseData{
        StatusCode: response.StatusCode,
        Headers:    response.Header.Clone(),
        Body:       body,
        ExpiresAt:  time.Now().Add(ttl),
    }
    s.data[key] = data
    return data, nil
}