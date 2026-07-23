package config

import (
	"strconv"
	"strings"
)

// yamlNode is a minimal recursive map used to represent parsed YAML.
// Only the subset of YAML that StreamPass config files actually use is
// supported (KISS/YAGNI): nested maps with 2-space indentation and scalar
// values (string/int/bool/float). Lists and anchors are intentionally not
// supported — no config file in this project needs them.
type yamlNode map[string]any

// yamlEntry is one parsed, non-blank YAML line before tree assembly.
type yamlEntry struct {
	indent   int
	key      string
	value    string
	hasValue bool
}

// parseYAML parses a minimal YAML document into a yamlNode tree.
//
// Whether a "key:" line with no trailing value starts a nested map or is
// simply an empty scalar (e.g. an unresolved ${VAR} placeholder that
// resolved to "") can't be decided from that line alone — it depends on
// whether a more-indented line follows. So parsing happens in two passes:
// first every non-blank line is collected into a flat list, then the tree
// is built with one line of lookahead to tell the two cases apart.
func parseYAML(content string) yamlNode {
	lines := strings.Split(content, "\n")
	var entries []yamlEntry
	for _, raw := range lines {
		line := stripComment(raw)
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := countIndent(line)
		trimmed := strings.TrimSpace(line)
		key, value, hasValue := splitKeyValue(trimmed)
		entries = append(entries, yamlEntry{indent: indent, key: key, value: value, hasValue: hasValue})
	}

	root := yamlNode{}
	stack := []struct {
		indent int
		node   yamlNode
	}{{indent: -1, node: root}}

	for i, e := range entries {
		// Pop stack until we find the parent for this indent level.
		for len(stack) > 1 && e.indent <= stack[len(stack)-1].indent {
			stack = stack[:len(stack)-1]
		}
		parent := stack[len(stack)-1].node

		nextIsChild := i+1 < len(entries) && entries[i+1].indent > e.indent
		if nextIsChild {
			child := yamlNode{}
			parent[e.key] = child
			stack = append(stack, struct {
				indent int
				node   yamlNode
			}{indent: e.indent, node: child})
			continue
		}

		if !e.hasValue || e.value == "" {
			// No children follow and no value was given: a genuinely empty
			// scalar (e.g. "key:" or a placeholder that resolved to ""),
			// not a nested map.
			parent[e.key] = ""
			continue
		}

		parent[e.key] = parseScalar(e.value)
	}

	return root
}

func stripComment(line string) string {
	inQuote := byte(0)
	for i := 0; i < len(line); i++ {
		c := line[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			inQuote = c
			continue
		}
		if c == '#' {
			return line[:i]
		}
	}
	return line
}

func countIndent(line string) int {
	n := 0
	for _, c := range line {
		if c == ' ' {
			n++
		} else {
			break
		}
	}
	return n
}

func splitKeyValue(trimmed string) (key, value string, hasValue bool) {
	idx := strings.Index(trimmed, ":")
	if idx == -1 {
		return trimmed, "", false
	}
	key = strings.TrimSpace(trimmed[:idx])
	value = strings.TrimSpace(trimmed[idx+1:])
	return key, value, true
}

func parseScalar(raw string) any {
	if len(raw) >= 2 && (raw[0] == '"' && raw[len(raw)-1] == '"' || raw[0] == '\'' && raw[len(raw)-1] == '\'') {
		return raw[1 : len(raw)-1]
	}
	if raw == "true" {
		return true
	}
	if raw == "false" {
		return false
	}
	if i, err := strconv.Atoi(raw); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		return f
	}
	return raw
}

// get navigates a dotted path (e.g. "database.host") through the tree.
func (n yamlNode) get(path string) (any, bool) {
	parts := strings.Split(path, ".")
	var cur any = n
	for _, p := range parts {
		m, ok := cur.(yamlNode)
		if !ok {
			return nil, false
		}
		v, ok := m[p]
		if !ok {
			return nil, false
		}
		cur = v
	}
	return cur, true
}
