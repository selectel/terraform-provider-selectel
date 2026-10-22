package schemadoc

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// Lint checks that the description of every attribute in a resource or data
// source schema states the effect of each of its rules. A schema without any
// schemadoc rule is not described by schemadoc, and Lint reports nothing for
// it. Otherwise it reports
//   - a validator, default or plan modifier without Documentation(), unless it
//     is runtime-only like UseStateForUnknown;
//   - facts missing from the description: the attribute uses a plain
//     Description, its kind is not supported by a describer, or rules changed
//     after the describer was called.
//
// The walk uses reflection over the exported schema fields, so it covers every
// attribute kind, nested attributes and blocks.
func Lint(schema any) []string {
	var l linter
	l.walk("", reflect.ValueOf(schema))
	if !l.usesSchemadoc {
		return nil
	}

	return l.problems
}

type linter struct {
	problems      []string
	usesSchemadoc bool
}

func (l *linter) walk(at string, v reflect.Value) {
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	if v.FieldByName("MarkdownDescription").IsValid() && at != "" {
		l.lintAttribute(at, v)
	}
	for _, children := range []reflect.Value{v.FieldByName("Attributes"), v.FieldByName("Blocks")} {
		if children.IsValid() && children.Kind() == reflect.Map {
			keys := children.MapKeys()
			slices.SortFunc(keys, func(a, b reflect.Value) int { return strings.Compare(a.String(), b.String()) })
			for _, key := range keys {
				l.walk(strings.TrimPrefix(at+"."+key.String(), "."), children.MapIndex(key))
			}
		}
	}
	if nested := v.FieldByName("NestedObject"); nested.IsValid() {
		l.walk(at, nested)
	}
}

func (l *linter) lintAttribute(at string, v reflect.Value) {
	description := v.FieldByName("MarkdownDescription").String()
	if description == "" {
		description = v.FieldByName("Description").String()
	}
	report := func(format string, args ...any) {
		l.problems = append(l.problems, at+": "+fmt.Sprintf(format, args...))
	}
	for _, rule := range rulesOf(v.Interface()) {
		documented, ok := rule.(Documented)
		if !ok {
			if !runtimeOnly(rule) {
				report("%T has no documentation facts; use a schemadoc rule or DescribeString", rule)
			}

			continue
		}
		l.usesSchemadoc = true
		facts := documented.Documentation()
		for _, text := range []string{facts.Constraint, facts.Dependency, facts.Default} {
			if text != "" && !strings.Contains(description, text) {
				report("description lacks %q", text)
			}
		}
		if facts.ReplacesResource && !strings.Contains(description, replacementPrefix) {
			report("description lacks the replacement sentence")
		}
		if facts.ReplacementException != "" && !strings.Contains(description, facts.ReplacementException) {
			report("description lacks %q", facts.ReplacementException)
		}
	}
}

// runtimeOnly rules do not change what a user can configure.
func runtimeOnly(rule any) bool {
	return strings.HasSuffix(fmt.Sprintf("%T", rule), ".useStateForUnknownModifier")
}
