package store

import (
	"fmt"
	"time"
)

var (
	DEFAULT_TTL = 120000 //120 seconds
)

type RedisMapValue struct {
	Value        string
	InsertedTime time.Time
}
type Store struct {
	RedisMap map[string]RedisMapValue
}

func New() *Store {
	return &Store{
		RedisMap: make(map[string]RedisMapValue),
	}
}

func (s *Store) Set(key, value string) {
	s.RedisMap[key] = RedisMapValue{
		Value:        value,
		InsertedTime: time.Now().Add(time.Millisecond * time.Duration(DEFAULT_TTL)),
	}
}

func (s *Store) SetWithTTL(key, value string, ttl int) {
	s.RedisMap[key] = RedisMapValue{
		Value:        value,
		InsertedTime: time.Now().Add(time.Millisecond * time.Duration(ttl)),
	}
}

func (s *Store) Get(key string) (string, bool) {
	if value, ok := s.RedisMap[key]; !ok {
		return "", false
	} else {
		if value.InsertedTime.IsZero() || time.Now().Before(value.InsertedTime) {
			return value.Value, true
		}
		delete(s.RedisMap, key)
		return "", false
	}
}
func (s *Store) PrintMap() {
	fmt.Println("Map Contents:")
	for key, value := range s.RedisMap {
		fmt.Printf("Key: %s  Value: %s", key, value)
	}
}
