package cache 

import (
	"testing"
	"time"
)

func TestGetExpiredEntryReturnsNil(t *testing.T) {
	store := NewStore()

	store.data["key"] = &ResponseData {
		StatusCode: 200,
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}

	result := store.Get("key")
	if result != nil {
		t.Errorf("expected nil for expired entry , got %+v", result)
	}
}

func TestGetValidEntryReturnsData(t *testing.T) {
    store := NewStore()

    expected := &ResponseData{
        StatusCode: 200,
        Body:       []byte("hello world"),
        ExpiresAt:  time.Now().Add(5 * time.Minute), 
    }

    store.data["key"] = expected

    result := store.Get("key")

    if result == nil {
        t.Fatal("expected a response, got nil")
    }
    if result.StatusCode != expected.StatusCode {
        t.Errorf("expected status %d, got %d", expected.StatusCode, result.StatusCode)
    }
    if string(result.Body) != string(expected.Body) {
        t.Errorf("expected body %q, got %q", expected.Body, result.Body)
    }
}