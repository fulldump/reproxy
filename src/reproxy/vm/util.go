package vm

import (
	"errors"
	"strconv"
)

func interface2number(i interface{}) (float64, error) {

	switch v := i.(type) {
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		if t, err := strconv.ParseFloat(v, 64); nil == err {
			return t, nil
		} else {
			return 0, errors.New("not float from string")
		}
	}

	return 0, errors.New("not float")
}
