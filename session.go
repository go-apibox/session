package session

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

type Session struct {
	sessions.Session
	r *http.Request
}

func (s *Session) Get(key string) (interface{}, error) {
	v, ok := s.Values[key]
	if !ok {
		return nil, errors.New("session key '" + key + "' not exist.")
	}
	return v, nil
}

func (s *Session) Set(key string, val interface{}) {
	s.Values[key] = val
}

func (s *Session) SetMap(m map[string]interface{}) {
	for k, v := range m {
		s.Values[k] = v
	}
}

func (s *Session) Delete(key string) {
	delete(s.Values, key)
}

func (s *Session) Save(w http.ResponseWriter) error {
	return s.Session.Save(s.r, w)
}

func (s *Session) Destroy(w http.ResponseWriter) error {
	s.Session.Options.MaxAge = -1
	return s.Session.Save(s.r, w)
}

func (s *Session) GetString(key string) (string, error) {
	v, err := s.Get(key)
	if err != nil {
		return "", err
	}

	if str, ok := v.(string); ok {
		return str, nil
	} else {
		return "", errors.New("session key '" + key + "' is not a string.")
	}
}

func (s *Session) GetBool(key string) (bool, error) {
	v, err := s.Get(key)
	if err != nil {
		return false, err
	}

	if b, ok := v.(bool); ok {
		return b, nil
	} else {
		return false, errors.New("session key '" + key + "' is not a boolean.")
	}
}

func (s *Session) GetInt(key string) (int, error) {
	v, err := s.Get(key)
	if err != nil {
		return 0, err
	}

	if b, ok := v.(int); ok {
		return b, nil
	} else {
		return 0, errors.New("session key '" + key + "' is not a int.")
	}
}

func (s *Session) GetInt32(key string) (int32, error) {
	v, err := s.Get(key)
	if err != nil {
		return 0, err
	}

	if b, ok := v.(int32); ok {
		return b, nil
	} else {
		return 0, errors.New("session key '" + key + "' is not a int32.")
	}
}

func (s *Session) GetInt64(key string) (int64, error) {
	v, err := s.Get(key)
	if err != nil {
		return 0, err
	}

	if b, ok := v.(int64); ok {
		return b, nil
	} else {
		return 0, errors.New("session key '" + key + "' is not a int64.")
	}
}

func (s *Session) GetUint(key string) (uint, error) {
	v, err := s.Get(key)
	if err != nil {
		return 0, err
	}

	if b, ok := v.(uint); ok {
		return b, nil
	} else {
		return 0, errors.New("session key '" + key + "' is not a uint.")
	}
}

func (s *Session) GetUint32(key string) (uint32, error) {
	v, err := s.Get(key)
	if err != nil {
		return 0, err
	}

	if b, ok := v.(uint32); ok {
		return b, nil
	} else {
		return 0, errors.New("session key '" + key + "' is not a uint32.")
	}
}

func (s *Session) GetUint64(key string) (uint64, error) {
	v, err := s.Get(key)
	if err != nil {
		return 0, err
	}

	if b, ok := v.(uint64); ok {
		return b, nil
	} else {
		return 0, errors.New("session key '" + key + "' is not a uint64.")
	}
}

func (s *Session) GetStringArray(key string) ([]string, error) {
	v, err := s.Get(key)
	if err != nil {
		return nil, err
	}

	if str, ok := v.([]string); ok {
		return str, nil
	} else {
		return nil, errors.New("session key '" + key + "' is not a string array.")
	}
}

func (s *Session) GetBoolArray(key string) ([]bool, error) {
	v, err := s.Get(key)
	if err != nil {
		return nil, err
	}

	if b, ok := v.([]bool); ok {
		return b, nil
	} else {
		return nil, errors.New("session key '" + key + "' is not a boolean array.")
	}
}

func (s *Session) GetIntArray(key string) ([]int, error) {
	v, err := s.Get(key)
	if err != nil {
		return nil, err
	}

	if b, ok := v.([]int); ok {
		return b, nil
	} else {
		return nil, errors.New("session key '" + key + "' is not a int array.")
	}
}

func (s *Session) GetInt32Array(key string) ([]int32, error) {
	v, err := s.Get(key)
	if err != nil {
		return nil, err
	}

	if b, ok := v.([]int32); ok {
		return b, nil
	} else {
		return nil, errors.New("session key '" + key + "' is not a int32 array.")
	}
}

func (s *Session) GetInt64Array(key string) ([]int64, error) {
	v, err := s.Get(key)
	if err != nil {
		return nil, err
	}

	if b, ok := v.([]int64); ok {
		return b, nil
	} else {
		return nil, errors.New("session key '" + key + "' is not a int64 array.")
	}
}

func (s *Session) GetUintArray(key string) ([]uint, error) {
	v, err := s.Get(key)
	if err != nil {
		return nil, err
	}

	if b, ok := v.([]uint); ok {
		return b, nil
	} else {
		return nil, errors.New("session key '" + key + "' is not a uint array.")
	}
}

func (s *Session) GetUint32Array(key string) ([]uint32, error) {
	v, err := s.Get(key)
	if err != nil {
		return nil, err
	}

	if b, ok := v.([]uint32); ok {
		return b, nil
	} else {
		return nil, errors.New("session key '" + key + "' is not a uint32 array.")
	}
}

func (s *Session) GetUint64Array(key string) ([]uint64, error) {
	v, err := s.Get(key)
	if err != nil {
		return nil, err
	}

	if b, ok := v.([]uint64); ok {
		return b, nil
	} else {
		return nil, errors.New("session key '" + key + "' is not a uint64 array.")
	}
}
