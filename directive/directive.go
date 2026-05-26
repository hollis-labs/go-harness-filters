package directive

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// Directive is a parsed @namespace:action key=value form.
//
// The parser preserves the original raw text on [Directive.Raw] so audit
// logs and policy decisions can show what the agent actually emitted,
// not just the structured interpretation.
type Directive struct {
	// Namespace is the part before the colon ("nanite", "torque",
	// "fragment", ...). Lowercased by the parser.
	Namespace string

	// Action is the part after the colon ("recover-session",
	// "block", "capture", ...). Lowercased; hyphens preserved.
	Action string

	// Args holds the key=value pairs in source order. Keys are
	// lowercased; values are decoded (quoted strings unquoted,
	// backslash escapes resolved). Order-preserving so callers can
	// detect duplicate keys.
	Args []Arg

	// Raw is the original directive text exactly as it appeared in
	// the source stream, leading "@" included. Used for audit and
	// disclosure.
	Raw string
}

// Arg is one key=value pair from a directive.
type Arg struct {
	Key   string
	Value string
}

// Get returns the first value bound to key, or "" if no such key exists.
// Callers that need to detect "no such key" vs "empty value" should walk
// [Directive.Args] directly.
func (d Directive) Get(key string) string {
	for _, a := range d.Args {
		if a.Key == key {
			return a.Value
		}
	}
	return ""
}

// Parse parses a single directive line. The leading "@" is required.
// Whitespace around the directive is tolerated; trailing text after the
// last argument is an error.
//
// Parse is intentionally strict — directives are control plane and
// silent malformed-directive handling would let agents submit
// near-misses that don't do what they look like.
func Parse(s string) (Directive, error) {
	raw := s
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "@") {
		return Directive{}, ErrMissingAtPrefix
	}
	s = s[1:]

	colonIdx := strings.IndexByte(s, ':')
	if colonIdx <= 0 || colonIdx == len(s)-1 {
		return Directive{}, fmt.Errorf("directive: malformed namespace:action in %q", raw)
	}
	ns := strings.ToLower(s[:colonIdx])
	rest := s[colonIdx+1:]

	if err := validateIdent(ns); err != nil {
		return Directive{}, fmt.Errorf("directive: namespace %q: %w", ns, err)
	}

	// Action runs until first whitespace.
	actionEnd := strings.IndexFunc(rest, unicode.IsSpace)
	var action, argsPart string
	if actionEnd < 0 {
		action = rest
		argsPart = ""
	} else {
		action = rest[:actionEnd]
		argsPart = rest[actionEnd:]
	}
	action = strings.ToLower(action)
	if err := validateAction(action); err != nil {
		return Directive{}, fmt.Errorf("directive: action %q: %w", action, err)
	}

	args, err := parseArgs(argsPart)
	if err != nil {
		return Directive{}, fmt.Errorf("directive: parsing args of %q: %w", raw, err)
	}

	return Directive{
		Namespace: ns,
		Action:    action,
		Args:      args,
		Raw:       strings.TrimSpace(raw),
	}, nil
}

// ErrMissingAtPrefix is returned when [Parse] receives input that does
// not start with "@". Use [errors.Is] to detect.
var ErrMissingAtPrefix = errors.New("directive: input does not start with @")

func validateIdent(s string) error {
	if s == "" {
		return errors.New("empty")
	}
	for _, r := range s {
		if !(unicode.IsLower(r) || unicode.IsDigit(r) || r == '-') {
			return fmt.Errorf("invalid character %q (allowed: a-z 0-9 -)", r)
		}
	}
	if s[0] == '-' || s[len(s)-1] == '-' {
		return errors.New("must not start or end with '-'")
	}
	return nil
}

func validateAction(s string) error { return validateIdent(s) }

// validateKey allows the same charset as validateIdent plus '_'. Keys
// commonly come from existing MCP-style arg conventions (e.g.,
// "short_code", "last_50") where underscores are idiomatic; namespaces
// and actions stay hyphen-only for URL-safety and slug-style consistency.
func validateKey(s string) error {
	if s == "" {
		return errors.New("empty")
	}
	for _, r := range s {
		if !(unicode.IsLower(r) || unicode.IsDigit(r) || r == '-' || r == '_') {
			return fmt.Errorf("invalid character %q (allowed: a-z 0-9 - _)", r)
		}
	}
	if s[0] == '-' || s[0] == '_' || s[len(s)-1] == '-' || s[len(s)-1] == '_' {
		return errors.New("must not start or end with '-' or '_'")
	}
	return nil
}

// parseArgs scans key=value pairs separated by whitespace. Values may be
// bare tokens (no whitespace/quotes/equals) or double-quoted strings
// with backslash escapes for \" and \\.
func parseArgs(s string) ([]Arg, error) {
	var out []Arg
	i := 0
	for i < len(s) {
		// Skip whitespace.
		for i < len(s) && isWS(s[i]) {
			i++
		}
		if i >= len(s) {
			break
		}

		// Read key up to '='.
		keyStart := i
		for i < len(s) && s[i] != '=' && !isWS(s[i]) {
			i++
		}
		if i >= len(s) || s[i] != '=' {
			return nil, fmt.Errorf("expected '=' after key starting at offset %d", keyStart)
		}
		key := strings.ToLower(s[keyStart:i])
		if err := validateKey(key); err != nil {
			return nil, fmt.Errorf("invalid key %q: %w", key, err)
		}
		i++ // consume '='

		// Read value.
		if i >= len(s) {
			return nil, fmt.Errorf("missing value for key %q", key)
		}
		var value string
		var err error
		if s[i] == '"' {
			value, i, err = readQuoted(s, i)
			if err != nil {
				return nil, fmt.Errorf("value for %q: %w", key, err)
			}
		} else {
			value, i = readBare(s, i)
		}
		out = append(out, Arg{Key: key, Value: value})
	}
	return out, nil
}

func isWS(b byte) bool { return b == ' ' || b == '\t' || b == '\n' || b == '\r' }

func readBare(s string, i int) (string, int) {
	start := i
	for i < len(s) && !isWS(s[i]) {
		i++
	}
	return s[start:i], i
}

func readQuoted(s string, i int) (string, int, error) {
	if s[i] != '"' {
		return "", i, errors.New("expected opening quote")
	}
	i++
	var sb strings.Builder
	for i < len(s) {
		c := s[i]
		switch c {
		case '"':
			return sb.String(), i + 1, nil
		case '\\':
			if i+1 >= len(s) {
				return "", i, errors.New("trailing backslash")
			}
			next := s[i+1]
			switch next {
			case '"', '\\':
				sb.WriteByte(next)
			default:
				return "", i, fmt.Errorf("unsupported escape \\%c (only \\\" and \\\\ are valid)", next)
			}
			i += 2
		default:
			sb.WriteByte(c)
			i++
		}
	}
	return "", i, errors.New("unterminated quoted value")
}
