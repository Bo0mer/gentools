package util

import (
	"regexp"
	"strings"
)

func SnakeCase(text string) string {
	result := text
	result = runRegExp("(.)([A-Z][a-z]+)", result, "${1}_${2}")
	result = runRegExp("([a-zA-Z])([0-9])", result, "${1}_${2}")
	result = runRegExp("([0-9])([a-zA-Z])", result, "${1}_${2}")
	result = runRegExp("([a-z])([A-Z])", result, "${1}_${2}")
	return strings.ToLower(result)
}

func runRegExp(expr string, text string, format string) string {
	model := regexp.MustCompile(expr)
	return model.ReplaceAllString(text, format)
}

func ToPrivate(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToLower(name[0:1]) + name[1:]
}
