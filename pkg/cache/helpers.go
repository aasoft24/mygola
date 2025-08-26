// pkg/cache/helpers.go
package cache

import (
	"encoding/json"
	"fmt"
	"time"
)

// GetJSON retrieves and unmarshals a JSON value from cache
func GetJSON(cache Cache, key string, v interface{}) error {
	data, err := cache.Get(key)
	if err != nil {
		return err
	}

	// Convert to JSON string
	jsonStr, ok := data.(string)
	if !ok {
		return fmt.Errorf("cached value is not a string")
	}

	// Unmarshal JSON
	return json.Unmarshal([]byte(jsonStr), v)
}

// SetJSON marshals a value to JSON and stores it in cache
func SetJSON(cache Cache, key string, v interface{}, expiration time.Duration) error {
	// Marshal to JSON
	jsonData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// Store as string
	return cache.Set(key, string(jsonData), expiration)
}
