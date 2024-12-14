package db

import (
	"database/sql/driver"
	"fmt"
	"slices"
	"strings"
)

// StringList represents a database type that stores a list of strings.
type StringList []string

// Scan implements [sql.Scanner].
func (sl *StringList) Scan(value any) error {
	if value == nil {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("scan StringList: %v", value)
	}

	parts := strings.Split(str, ";")
	*sl = StringList(parts)
	return nil
}

// Value implements [sql.Valuer].
func (sl StringList) Value() (driver.Value, error) {
	if len(sl) == 0 {
		return nil, nil
	}

	return strings.Join(sl, ";"), nil
}

const (
	stringMapKVSep   = "==="
	stringMapPairSep = "@@@"
)

// StringMap represents a field that contains multiple key/value pairs.
type StringMap map[string]string

// Scan implements [sql.Scanner].
func (sm *StringMap) Scan(value any) error {
	if value == nil {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("scan StringMap: %v", value)
	}

	parts := strings.Split(str, stringMapPairSep)
	m := make(map[string]string)
	for _, part := range parts {
		keyVal := strings.Split(part, stringMapKVSep)
		if len(keyVal) != 2 {
			return fmt.Errorf("map item has wrong format: %v", part)
		}

		m[keyVal[0]] = keyVal[1]
	}
	*sm = StringMap(m)
	return nil
}

// String serializes the content of the map to a string.
func (sl StringMap) String() string {
	if len(sl) == 0 {
		return ""
	}

	keys := []string{}
	for k := range sl {
		keys = append(keys, k)
	}

	slices.Sort(keys)
	var s string
	for idx, k := range keys {
		s += k + stringMapKVSep + sl[k]
		if idx < (len(keys) - 1) {
			s += stringMapPairSep
		}
	}

	return s
}

// Value implements [sql.Valuer].
func (sl StringMap) Value() (driver.Value, error) {
	if len(sl) == 0 {
		return nil, nil
	}

	return sl.String(), nil
}
