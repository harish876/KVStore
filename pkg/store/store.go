package store

type Store struct {
	RedisMap map[string]string
}

func New() *Store {
	return &Store{
		RedisMap: make(map[string]string),
	}
}

func (s *Store) Set(key, value string) {
	s.RedisMap[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	if value, ok := s.RedisMap[key]; !ok {
		return "", false
	} else {
		return value, true
	}
}
