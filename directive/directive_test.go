package directive

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseValidDirectives(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Directive
	}{
		{
			name: "namespace:action only",
			in:   "@torque:status",
			want: Directive{
				Namespace: "torque", Action: "status",
				Raw: "@torque:status",
			},
		},
		{
			name: "bare args",
			in:   "@nanite:recover-session short_code=c287 history=last_50",
			want: Directive{
				Namespace: "nanite", Action: "recover-session",
				Args: []Arg{
					{Key: "short_code", Value: "c287"},
					{Key: "history", Value: "last_50"},
				},
				Raw: "@nanite:recover-session short_code=c287 history=last_50",
			},
		},
		{
			name: "quoted value with spaces",
			in:   `@fragment:capture kind=decision title="Use Cerberus for Nanite deploys"`,
			want: Directive{
				Namespace: "fragment", Action: "capture",
				Args: []Arg{
					{Key: "kind", Value: "decision"},
					{Key: "title", Value: "Use Cerberus for Nanite deploys"},
				},
				Raw: `@fragment:capture kind=decision title="Use Cerberus for Nanite deploys"`,
			},
		},
		{
			name: "escaped quote in value",
			in:   `@x:y msg="he said \"hi\""`,
			want: Directive{
				Namespace: "x", Action: "y",
				Args: []Arg{{Key: "msg", Value: `he said "hi"`}},
				Raw:  `@x:y msg="he said \"hi\""`,
			},
		},
		{
			name: "uppercase namespace+action+key normalized",
			in:   "@TORQUE:BLOCK Task=foo",
			want: Directive{
				Namespace: "torque", Action: "block",
				Args: []Arg{{Key: "task", Value: "foo"}},
				Raw:  "@TORQUE:BLOCK Task=foo",
			},
		},
		{
			name: "leading and trailing whitespace tolerated",
			in:   "   @torque:status   ",
			want: Directive{
				Namespace: "torque", Action: "status",
				Raw: "@torque:status",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Parse(c.in)
			if err != nil {
				t.Fatalf("Parse(%q): %v", c.in, err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("Parse(%q):\n  got  %#v\n  want %#v", c.in, got, c.want)
			}
		})
	}
}

func TestParseRejectsMalformed(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"no @", "torque:status"},
		{"empty namespace", "@:status"},
		{"empty action", "@torque:"},
		{"missing colon", "@torque status"},
		{"invalid namespace char", "@tor!que:status"},
		{"key without =", "@torque:status onlyflag"},
		{"unterminated quote", `@x:y k="oops`},
		{"unknown escape", `@x:y k="\n"`},
		{"missing value after =", "@x:y k="},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Parse(c.in); err == nil {
				t.Fatalf("Parse(%q) succeeded; want error", c.in)
			}
		})
	}
}

func TestErrMissingAtPrefixIsSentinel(t *testing.T) {
	_, err := Parse("torque:status")
	if !errors.Is(err, ErrMissingAtPrefix) {
		t.Fatalf("err = %v, want errors.Is(err, ErrMissingAtPrefix)", err)
	}
}

func TestDirectiveGet(t *testing.T) {
	d := Directive{Args: []Arg{
		{Key: "a", Value: "1"},
		{Key: "b", Value: "2"},
		{Key: "a", Value: "shadowed"}, // duplicates: Get returns first
	}}
	if got := d.Get("a"); got != "1" {
		t.Errorf("Get(a) = %q, want 1", got)
	}
	if got := d.Get("b"); got != "2" {
		t.Errorf("Get(b) = %q, want 2", got)
	}
	if got := d.Get("missing"); got != "" {
		t.Errorf("Get(missing) = %q, want empty string", got)
	}
}
