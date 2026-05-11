package repository

import "testing"

func TestApplyTemplate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		in     string
		params map[string]string
		want   string
	}{
		{
			name:   "empty input returns empty",
			in:     "",
			params: map[string]string{"id": "42"},
			want:   "",
		},
		{
			name:   "no params returns input unchanged",
			in:     "{{id}}",
			params: nil,
			want:   "{{id}}",
		},
		{
			name:   "single substitution",
			in:     `{"id":{{id}}}`,
			params: map[string]string{"id": "42"},
			want:   `{"id":42}`,
		},
		{
			name:   "string substitution preserves quotes",
			in:     `{"id":"{{id}}"}`,
			params: map[string]string{"id": "42"},
			want:   `{"id":"42"}`,
		},
		{
			name:   "multiple tokens",
			in:     `org={{org}} user={{user}}`,
			params: map[string]string{"org": "acme", "user": "ada"},
			want:   `org=acme user=ada`,
		},
		{
			name:   "same token repeated",
			in:     `{{id}}-{{id}}-{{id}}`,
			params: map[string]string{"id": "x"},
			want:   `x-x-x`,
		},
		{
			name:   "whitespace inside braces is tolerated",
			in:     `{{ id }} and {{  user  }}`,
			params: map[string]string{"id": "1", "user": "ada"},
			want:   `1 and ada`,
		},
		{
			name:   "unknown token left as-is",
			in:     `{{known}} and {{unknown}}`,
			params: map[string]string{"known": "yes"},
			want:   `yes and {{unknown}}`,
		},
		{
			name:   "single braces are not tokens",
			in:     `{"id":"{not-a-token}"}`,
			params: map[string]string{"id": "42"},
			want:   `{"id":"{not-a-token}"}`,
		},
		{
			name:   "token at start and end",
			in:     `{{a}}-middle-{{b}}`,
			params: map[string]string{"a": "A", "b": "B"},
			want:   `A-middle-B`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ApplyTemplate(tc.in, tc.params)
			if got != tc.want {
				t.Errorf("ApplyTemplate(%q, %v) = %q, want %q", tc.in, tc.params, got, tc.want)
			}
		})
	}
}
