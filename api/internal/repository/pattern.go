package repository

import (
	"fmt"
	"regexp"
	"strings"
)

// Supported path-parameter types. Each maps to a regex fragment used to match
// a single path segment of the declared shape.
var pathParamTypeRegex = map[string]string{
	"string": `[^/]+`,
	"int":    `-?\d+`,
	"uuid":   `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`,
}

// SupportedPathParamTypes returns the type names accepted in a mock's
// pathParams map.
func SupportedPathParamTypes() []string {
	out := make([]string, 0, len(pathParamTypeRegex))
	for k := range pathParamTypeRegex {
		out = append(out, k)
	}
	return out
}

// IsSupportedPathParamType reports whether t is one of the supported type
// names (case-sensitive).
func IsSupportedPathParamType(t string) bool {
	_, ok := pathParamTypeRegex[t]
	return ok
}

// HasPatternSyntax reports whether path contains a "{name}" token.
func HasPatternSyntax(path string) bool {
	return strings.Contains(path, "{")
}

// ExtractPathParamNames returns the parameter names declared in path, in
// order. Returns an error for empty or duplicate names, or for unmatched
// braces.
func ExtractPathParamNames(path string) ([]string, error) {
	var names []string
	seen := make(map[string]struct{})

	runes := []rune(path)
	for i := 0; i < len(runes); i++ {
		if runes[i] != '{' {
			continue
		}
		end := -1
		for j := i + 1; j < len(runes); j++ {
			if runes[j] == '}' {
				end = j
				break
			}
		}
		if end == -1 {
			return nil, fmt.Errorf("unmatched { in path %q", path)
		}
		name := string(runes[i+1 : end])
		if name == "" {
			return nil, fmt.Errorf("empty path param name in %q", path)
		}
		if strings.ContainsAny(name, "{}/") {
			return nil, fmt.Errorf("invalid path param name %q", name)
		}
		if _, dup := seen[name]; dup {
			return nil, fmt.Errorf("duplicate path param name %q", name)
		}
		seen[name] = struct{}{}
		names = append(names, name)
		i = end
	}
	return names, nil
}

// PatternToRegex builds an anchored regex from a pattern path and its declared
// param types. Each "{name}" becomes a named capture group "(?P<name>seg)" so
// captured values can be retrieved by name via ExtractCapturedValues.
//
// Returns nil if the path is malformed, references an unknown param name, or
// uses an unsupported type — callers should skip such entries during lookup.
func PatternToRegex(path string, params map[string]string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")

	runes := []rune(path)
	for i := 0; i < len(runes); {
		c := runes[i]
		if c == '{' {
			end := -1
			for j := i + 1; j < len(runes); j++ {
				if runes[j] == '}' {
					end = j
					break
				}
			}
			if end == -1 {
				return nil
			}
			name := string(runes[i+1 : end])
			typeName, ok := params[name]
			if !ok {
				return nil
			}
			seg, ok := pathParamTypeRegex[typeName]
			if !ok {
				return nil
			}
			b.WriteString("(?P<")
			b.WriteString(name)
			b.WriteString(">")
			b.WriteString(seg)
			b.WriteString(")")
			i = end + 1
			continue
		}
		b.WriteString(regexp.QuoteMeta(string(c)))
		i++
	}
	b.WriteString("$")

	re, err := regexp.Compile(b.String())
	if err != nil {
		return nil
	}
	return re
}

// ExtractCapturedValues runs re against path and returns a map of named
// capture name → captured value. Returns nil if path does not match.
func ExtractCapturedValues(re *regexp.Regexp, path string) map[string]string {
	match := re.FindStringSubmatch(path)
	if match == nil {
		return nil
	}
	names := re.SubexpNames()
	out := make(map[string]string, len(match)-1)
	for i, name := range names {
		if i == 0 || name == "" {
			continue
		}
		out[name] = match[i]
	}
	return out
}
