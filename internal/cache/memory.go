package cache

import (
	"io"
	"net/http"
	"sync"
	"time"
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

    defer response.Body.Close()
    body, err := io.ReadAll(response.Body)
    if err != nil {
        return nil, err
    }
   
    data := &ResponseData{
        StatusCode: response.StatusCode,
        Headers:    response.Header.Clone(),
        Body:       body,
        ExpiresAt:  time.Now().Add(ttl),
    }
    s.data[key] = data
    return data, nil
}

func (s *Store) StartCleanup(interval time.Duration) {
    go func () {
        ticker :=  time.NewTicker(interval)
        defer ticker.Stop()
        for range ticker.C {
            s.purgeExpired()
        }
    }()
}

func (s *Store) purgeExpired() {
    s.mu.Lock()
    defer s.mu.Unlock()
    for key, entry := range s.data {
        if  time.Now().After(entry.ExpiresAt) {
            delete(s.data,  key)
        }
    }
}

func (s *Store) Clear() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data = make(map[string]*ResponseData) 
}