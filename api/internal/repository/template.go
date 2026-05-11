package repository

import "regexp"

// templateTokenRe matches "{{ name }}" where name is a simple identifier.
// Whitespace around the name is allowed and ignored.
var templateTokenRe = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)

// ApplyTemplate replaces every "{{name}}" token in s with the value at
// params[name]. Tokens whose name is not in params are left unchanged so
// they're easy to spot in the response during debugging.
func ApplyTemplate(s string, params map[string]string) string {
	if s == "" || len(params) == 0 {
		return s
	}
	return templateTokenRe.ReplaceAllStringFunc(s, func(token string) string {
		sub := templateTokenRe.FindStringSubmatch(token)
		if sub == nil {
			return token
		}
		if v, ok := params[sub[1]]; ok {
			return v
		}
		return token
	})
}
