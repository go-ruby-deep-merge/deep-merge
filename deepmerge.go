// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-deep-merge/deep-merge authors

package deepmerge

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Options mirrors the option set of the Ruby deep_merge gem. The zero value
// behaves like Ruby's deep_merge! (overwrite unmergeable/conflicting values).
type Options struct {
	// Overwrite forces overwriting of unmergeable/conflicting values. Together
	// with PreserveUnmergeables it selects the gem's bang vs non-bang variant:
	// the effective flag is Overwrite || !PreserveUnmergeables.
	Overwrite bool

	// PreserveUnmergeables keeps existing dest values when a value cannot be
	// merged (type mismatch), matching Ruby's deep_merge (non-bang).
	PreserveUnmergeables bool

	// KnockoutPrefix, when non-empty, enables knockout: a source string element
	// prefixed with it removes the matching element from dest.
	KnockoutPrefix string

	// OverwriteArrays replaces dest arrays with source arrays instead of merging
	// them (the gem's :overwrite_arrays).
	OverwriteArrays bool

	// SortMergedArrays sorts every array produced by a merge.
	SortMergedArrays bool

	// UnpackArrays, when non-empty, joins then splits every merged array on this
	// separator before merging, unpacking compound string elements.
	UnpackArrays string

	// MergeHashArrays merges arrays of hashes element-wise (by index) when both
	// the source and dest arrays contain only hashes.
	MergeHashArrays bool

	// ExtendExistingArrays pushes a source hash/scalar onto a dest array rather
	// than overwriting the array.
	ExtendExistingArrays bool

	// KeepArrayDuplicates concatenates arrays keeping duplicate elements instead
	// of taking their set union.
	KeepArrayDuplicates bool

	// MergeNilValues merges nil source values instead of skipping them.
	MergeNilValues bool
}

// InvalidOptionError is the value passed to panic when Options combine in a way
// the deep_merge gem rejects.
type InvalidOptionError struct{ Message string }

func (e *InvalidOptionError) Error() string { return e.Message }

// config holds the resolved, immutable settings for one merge run.
type config struct {
	overwriteUnmergeable bool
	hasKnockout          bool
	knockoutPrefix       string
	hasUnpack            bool
	unpack               string
	overwriteArrays      bool
	sortMergedArrays     bool
	mergeHashArrays      bool
	extendExistingArrays bool
	keepArrayDuplicates  bool
	mergeNilValues       bool
}

func (o Options) resolve() *config {
	overwrite := o.Overwrite || !o.PreserveUnmergeables
	if o.KnockoutPrefix != "" && !overwrite {
		panic(&InvalidOptionError{Message: "deep-merge: KnockoutPrefix requires overwrite; set Overwrite or clear PreserveUnmergeables"})
	}
	return &config{
		overwriteUnmergeable: overwrite,
		hasKnockout:          o.KnockoutPrefix != "",
		knockoutPrefix:       o.KnockoutPrefix,
		hasUnpack:            o.UnpackArrays != "",
		unpack:               o.UnpackArrays,
		overwriteArrays:      o.OverwriteArrays,
		sortMergedArrays:     o.SortMergedArrays,
		mergeHashArrays:      o.MergeHashArrays,
		extendExistingArrays: o.ExtendExistingArrays,
		keepArrayDuplicates:  o.KeepArrayDuplicates,
		mergeNilValues:       o.MergeNilValues,
	}
}

// DeepMerge merges source into a deep copy of dest and returns the copy. The
// caller's dest is not modified.
func DeepMerge(dest, source any, opts Options) any {
	return merge(source, deepCopy(dest), opts.resolve())
}

// DeepMergeInto merges source into dest in place and returns the result. The
// return value must be used because the top-level value may be replaced.
func DeepMergeInto(dest, source any, opts Options) any {
	return merge(source, dest, opts.resolve())
}

// merge is the recursive core, faithful to DeepMerge::deep_merge!(source, dest).
func merge(source, dest any, c *config) any {
	if !c.mergeNilValues && source == nil {
		return dest
	}
	if !truthy(dest) && c.overwriteUnmergeable {
		return deepCopy(source)
	}

	switch src := source.(type) {
	case map[string]any:
		if dm, ok := dest.(map[string]any); ok {
			for _, k := range sortedKeys(src) {
				sv := src[k]
				if existing, ok := dm[k]; ok && truthy(existing) {
					dm[k] = merge(sv, existing, c)
					continue
				}
				// The key is absent (or nil/false): create it by merging the
				// source value into a fresh copy so knockout/unpack still apply.
				if _, isArr := sv.([]any); isArr && c.keepArrayDuplicates {
					dm[k] = merge(sv, []any{}, c)
				} else {
					dm[k] = merge(sv, deepCopy(sv), c)
				}
			}
			return dm
		}
		if da, ok := dest.([]any); ok && c.extendExistingArrays {
			return push(da, source)
		}
		if c.overwriteUnmergeable {
			return overwriteUnmergeables(source, dest, c)
		}
		return dest
	case []any:
		return mergeArray(src, dest, c)
	default:
		if da, ok := dest.([]any); ok && c.extendExistingArrays {
			return push(da, source)
		}
		return overwriteUnmergeables(source, dest, c)
	}
}

// mergeArray merges an array source into dest.
func mergeArray(source []any, dest any, c *config) any {
	if c.overwriteArrays {
		return deepCopy(source)
	}
	if c.hasUnpack {
		source = unpackArray(source, c.unpack)
		if da, ok := dest.([]any); ok {
			dest = unpackArray(da, c.unpack)
		}
	}
	// A bare knockout token anywhere in source truncates dest entirely.
	if c.hasKnockout && containsValue(source, c.knockoutPrefix) {
		dest = clearOrNil(dest)
		source = removeValue(source, c.knockoutPrefix)
	}
	if da, ok := dest.([]any); ok {
		if c.hasKnockout {
			source, da = knockout(source, da, c.knockoutPrefix)
		}
		var merged []any
		switch {
		case c.mergeHashArrays && allHashes(source) && allHashes(da):
			merged = make([]any, 0, max(len(source), len(da)))
			for i := range da {
				var s any = map[string]any{}
				if i < len(source) {
					s = source[i]
				}
				merged = append(merged, merge(s, da[i], c))
			}
			for _, v := range source[min(len(da), len(source)):] {
				merged = append(merged, deepCopy(v))
			}
		case c.keepArrayDuplicates:
			merged = concat(da, source)
		default:
			merged = union(da, source)
		}
		if c.sortMergedArrays {
			sortArray(merged)
		}
		return merged
	}
	if c.overwriteUnmergeable {
		return overwriteUnmergeables(source, dest, c)
	}
	return dest
}

// overwriteUnmergeables reproduces DeepMerge::overwrite_unmergeables.
func overwriteUnmergeables(source, dest any, c *config) any {
	if c.hasKnockout && c.overwriteUnmergeable {
		switch s := source.(type) {
		case string:
			stripped := strings.TrimPrefix(s, c.knockoutPrefix)
			if stripped == s {
				return stripped // no knockout prefix: overwrite dest with source
			}
			return "" // knockout prefix present: delete dest
		case []any:
			return removeKnockoutElems(s, c.knockoutPrefix)
		default:
			return deepCopy(source)
		}
	}
	if c.overwriteUnmergeable {
		return deepCopy(source)
	}
	return dest
}

// knockout removes knockout-prefixed elements from source and the matching plain
// (and prefixed) values from dest, returning the filtered pair.
func knockout(source, dest []any, prefix string) ([]any, []any) {
	newSource := make([]any, 0, len(source))
	d := append([]any{}, dest...)
	for _, ko := range source {
		if s, ok := ko.(string); ok && strings.HasPrefix(s, prefix) {
			item := strings.TrimPrefix(s, prefix)
			d = removeValue(d, item)
			d = removeValue(d, s)
			continue
		}
		newSource = append(newSource, ko)
	}
	return newSource, d
}

// removeKnockoutElems drops every string element of source that starts with the
// knockout prefix, mirroring the array branch of overwrite_unmergeables.
func removeKnockoutElems(source []any, prefix string) []any {
	out := make([]any, 0, len(source))
	for _, v := range source {
		if s, ok := v.(string); ok && strings.HasPrefix(s, prefix) {
			continue
		}
		out = append(out, v)
	}
	return out
}

// unpackArray joins then splits an array on sep (Ruby Array#join / String#split).
func unpackArray(a []any, sep string) []any {
	return rubySplit(strings.Join(toStrings(a), sep), sep)
}

func toStrings(a []any) []string {
	out := make([]string, len(a))
	for i, v := range a {
		out[i] = toRubyString(v)
	}
	return out
}

func toRubyString(v any) string {
	switch s := v.(type) {
	case nil:
		return ""
	case string:
		return s
	default:
		return fmt.Sprint(v)
	}
}

// rubySplit splits s on sep with Ruby's default String#split semantics: an empty
// string yields no fields and trailing empty fields are dropped.
func rubySplit(s, sep string) []any {
	if s == "" {
		return []any{}
	}
	parts := strings.Split(s, sep)
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	out := make([]any, len(parts))
	for i, p := range parts {
		out[i] = p
	}
	return out
}

// union returns the set union dest|source: dest elements first, then source
// elements not already present, with structural duplicates removed.
func union(a, b []any) []any {
	out := make([]any, 0, len(a)+len(b))
	add := func(x any) {
		for _, y := range out {
			if valuesEqual(x, y) {
				return
			}
		}
		out = append(out, x)
	}
	for _, x := range a {
		add(x)
	}
	for _, x := range b {
		add(x)
	}
	return out
}

// concat appends source to dest keeping all duplicates.
func concat(a, b []any) []any {
	out := make([]any, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

// push returns a new array with v appended to a copy of a.
func push(a []any, v any) []any {
	out := append([]any{}, a...)
	return append(out, deepCopy(v))
}

func containsValue(a []any, v any) bool {
	for _, x := range a {
		if valuesEqual(x, v) {
			return true
		}
	}
	return false
}

func removeValue(a []any, v any) []any {
	out := make([]any, 0, len(a))
	for _, x := range a {
		if !valuesEqual(x, v) {
			out = append(out, x)
		}
	}
	return out
}

func allHashes(a []any) bool {
	for _, v := range a {
		if _, ok := v.(map[string]any); !ok {
			return false
		}
	}
	return true
}

// clearOrNil empties a clearable dest (array, hash, string) or nils anything
// else, matching DeepMerge::clear_or_nil.
func clearOrNil(v any) any {
	switch v.(type) {
	case []any:
		return []any{}
	case map[string]any:
		return map[string]any{}
	case string:
		return ""
	default:
		return nil
	}
}

func sortArray(a []any) {
	sort.SliceStable(a, func(i, j int) bool {
		return compareValues(a[i], a[j]) < 0
	})
}

func compareValues(a, b any) int {
	if fa, ok := asFloat(a); ok {
		if fb, ok := asFloat(b); ok {
			switch {
			case fa < fb:
				return -1
			case fa > fb:
				return 1
			default:
				return 0
			}
		}
	}
	sa, aok := a.(string)
	sb, bok := b.(string)
	if aok && bok {
		return strings.Compare(sa, sb)
	}
	return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
}

func asFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

// truthy reports Ruby truthiness: everything except nil and false is truthy.
func truthy(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return true
}

func valuesEqual(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// deepCopy returns an independent copy of v; scalars are returned unchanged.
func deepCopy(v any) any {
	switch t := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[k] = deepCopy(val)
		}
		return m
	case []any:
		s := make([]any, len(t))
		for i, val := range t {
			s[i] = deepCopy(val)
		}
		return s
	default:
		return v
	}
}
