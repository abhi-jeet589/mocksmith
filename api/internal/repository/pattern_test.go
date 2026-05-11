package repository

import (
	"testing"
)

func TestHasPatternSyntax(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path string
		want bool
	}{
		{"/users/42", false},
		{"/", false},
		{"", false},
		{"/users/{id}", true},
		{"/{a}/{b}", true},
		{"/users/{id", true}, // unmatched, but pattern syntax is present
		{"/no/special/chars", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			if got := HasPatternSyntax(tc.path); got != tc.want {
				t.Errorf("HasPatternSyntax(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestIsSupportedPathParamType(t *testing.T) {
	t.Parallel()
	cases := []struct {
		typ  string
		want bool
	}{
		{"string", true},
		{"int", true},
		{"uuid", true},
		{"", false},
		{"INT", false}, // case sensitive
		{"float", false},
		{"bool", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.typ, func(t *testing.T) {
			t.Parallel()
			if got := IsSupportedPathParamType(tc.typ); got != tc.want {
				t.Errorf("IsSupportedPathParamType(%q) = %v, want %v", tc.typ, got, tc.want)
			}
		})
	}
}

func TestExtractPathParamNames(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		path      string
		wantNames []string
		wantErr   bool
	}{
		{"no params", "/users/42", nil, false},
		{"empty path", "", nil, false},
		{"single param", "/users/{id}", []string{"id"}, false},
		{"multiple params", "/orgs/{org}/users/{userId}", []string{"org", "userId"}, false},
		{"adjacent params", "/{a}{b}", []string{"a", "b"}, false},
		{"unmatched brace", "/users/{id", nil, true},
		{"empty name", "/users/{}", nil, true},
		{"duplicate name", "/{id}/{id}", nil, true},
		{"slash in name", "/users/{a/b}", nil, true},
		{"nested brace in name", "/users/{a{b}}", nil, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ExtractPathParamNames(tc.path)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ExtractPathParamNames(%q) error = %v, wantErr %v", tc.path, err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if !equalSlices(got, tc.wantNames) {
				t.Errorf("ExtractPathParamNames(%q) = %v, want %v", tc.path, got, tc.wantNames)
			}
		})
	}
}

func TestPatternToRegex_Matching(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		path      string
		params    map[string]string
		matches   []string
		noMatches []string
	}{
		{
			name:   "string param matches any single segment",
			path:   "/users/{id}",
			params: map[string]string{"id": "string"},
			matches: []string{
				"/users/42",
				"/users/abc",
				"/users/abc-def_123",
				"/users/has.dots",
			},
			noMatches: []string{
				"/users/",
				"/users/a/b",       // segment with slash
				"/users/42/extra",  // extra segment
				"/users",           // missing segment
				"/USERS/42",        // case sensitive literal
			},
		},
		{
			name:   "int param matches digits only",
			path:   "/users/{id}",
			params: map[string]string{"id": "int"},
			matches: []string{
				"/users/0",
				"/users/42",
				"/users/-7",
				"/users/1234567890",
			},
			noMatches: []string{
				"/users/abc",
				"/users/12abc",
				"/users/4.2",
				"/users/",
			},
		},
		{
			name:   "uuid param matches canonical 8-4-4-4-12",
			path:   "/items/{id}",
			params: map[string]string{"id": "uuid"},
			matches: []string{
				"/items/550e8400-e29b-41d4-a716-446655440000",
				"/items/00000000-0000-0000-0000-000000000000",
				"/items/AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE",
			},
			noMatches: []string{
				"/items/550e8400-e29b-41d4-a716-44665544000",   // 11 char last group
				"/items/notauuid",
				"/items/550e8400e29b41d4a716446655440000",        // no dashes
				"/items/550e8400-e29b-41d4-a716-446655440000-x",  // trailing junk
			},
		},
		{
			name:   "multiple params of mixed types",
			path:   "/orgs/{org}/users/{userId}",
			params: map[string]string{"org": "string", "userId": "int"},
			matches: []string{
				"/orgs/acme/users/42",
				"/orgs/some-org/users/0",
			},
			noMatches: []string{
				"/orgs/acme/users/abc", // userId must be int
				"/orgs/a/b/users/42",   // org segment can't span slashes
				"/orgs/acme/users/",
			},
		},
		{
			name:    "literal path with no params",
			path:    "/health",
			params:  nil,
			matches: []string{"/health"},
			noMatches: []string{
				"/HEALTH",
				"/health/",
				"/health/extra",
			},
		},
		{
			name:   "path with regex metacharacters is escaped",
			path:   "/v1/users.json/{id}",
			params: map[string]string{"id": "int"},
			matches: []string{
				"/v1/users.json/42",
			},
			noMatches: []string{
				"/v1/usersAjson/42", // '.' should be literal, not match any char
				"/v1/users.json/abc",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			re := PatternToRegex(tc.path, tc.params)
			if re == nil {
				t.Fatalf("PatternToRegex(%q, %v) returned nil", tc.path, tc.params)
			}
			for _, m := range tc.matches {
				if !re.MatchString(m) {
					t.Errorf("expected %q to match pattern %q (regex %q)", m, tc.path, re.String())
				}
			}
			for _, m := range tc.noMatches {
				if re.MatchString(m) {
					t.Errorf("expected %q NOT to match pattern %q (regex %q)", m, tc.path, re.String())
				}
			}
		})
	}
}

func TestPatternToRegex_ReturnsNil(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		path   string
		params map[string]string
	}{
		{
			name:   "unmatched brace",
			path:   "/users/{id",
			params: map[string]string{"id": "int"},
		},
		{
			name:   "param missing from declared map",
			path:   "/users/{id}",
			params: map[string]string{},
		},
		{
			name:   "param has unknown type",
			path:   "/users/{id}",
			params: map[string]string{"id": "octopus"},
		},
		{
			name:   "param has empty type",
			path:   "/users/{id}",
			params: map[string]string{"id": ""},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if re := PatternToRegex(tc.path, tc.params); re != nil {
				t.Errorf("expected nil regex for %q with %v, got %q", tc.path, tc.params, re.String())
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
