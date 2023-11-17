package env

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Type interface {
	bool | string | int | []string | []int | time.Duration
}

func Lookup[T Type](key string, defaultValue T) T {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	var res any
	switch any(defaultValue).(type) {
	case bool:
		value = strings.ToLower(value)
		valid := value == "1" || value == "true" || value == "ok" || value == "yes"
		if value == "" || !valid {
			return defaultValue
		}

		res = valid
	case time.Duration:
		d, err := time.ParseDuration(value)
		if err != nil {
			return defaultValue
		}

		res = d
	case int:
		i, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return defaultValue
		}

		res = int(i)
	case []int:
		if value == "" {
			return defaultValue
		}

		list := strings.Split(value, ",")
		arr := make([]int, 0, len(list))

		for _, v := range list {
			i, err := strconv.ParseInt(v, 10, 0)
			if err != nil {
				return defaultValue
			}
			arr = append(arr, int(i))
		}

		res = arr
	case string:
		if value == "" {
			return defaultValue
		}

		res = value
	case []string:
		if value == "" {
			return defaultValue
		}

		res = strings.Split(value, ",")
	}

	return res.(T)
}
