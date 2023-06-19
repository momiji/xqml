package xqml

import (
	"encoding/json"
)

func Stringify(v any) string {
	s, _ := json.Marshal(v)
	return string(s)
}

func ToJson(b []byte) (map[string]any, error) {
	var v map[string]any
	err := json.Unmarshal(b, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}
