package config

import (
	"strconv"
	"strings"
)

func IntOr(num string, fallback int) int {
	parsed, err := strconv.Atoi(num)
	if err != nil {
		return fallback
	}
	return parsed
}

func StringListOr(list string, fallback []string) []string {
	if len(list) == 0 {
		return fallback
	}
	return strings.Split(list, ",")
}
