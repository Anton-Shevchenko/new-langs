package helper

import (
	"fmt"
	"strconv"
	"strings"
)

const DefaultSeparator = "||"

func DataToString(elements ...interface{}) string {
	var result []string
	for _, element := range elements {
		switch v := element.(type) {
		case string:
			result = append(result, v)
		case int, float64, int64, float32:
			result = append(result, fmt.Sprintf("%v", v))
		default:
			result = append(result, fmt.Sprintf("%v", v))
		}
	}
	return strings.Join(result, DefaultSeparator)
}

func StringToData(input string) []interface{} {
	parts := strings.Split(input, DefaultSeparator)
	var result []interface{}

	for _, part := range parts {
		if num, err := strconv.Atoi(part); err == nil {
			result = append(result, num)
		} else if num, err := strconv.ParseFloat(part, 64); err == nil {
			result = append(result, num)
		} else {
			result = append(result, part)
		}
	}

	return result
}
