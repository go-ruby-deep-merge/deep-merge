// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-deep-merge/deep-merge authors

// Package deepmerge is a pure-Go (no cgo) port of the Ruby deep_merge gem
// (https://github.com/danielsdeleo/deep_merge). It recursively merges values
// that model Ruby data: hashes as map[string]any, arrays as []any, and any
// other Go value as a scalar.
//
// # Entry points
//
//   - [DeepMerge] merges source into a deep copy of dest and returns the copy;
//     the caller's dest is left untouched (non-destructive).
//   - [DeepMergeInto] merges source into dest in place and returns the result.
//     Because the top-level value may be replaced (for example when dest is nil
//     or a scalar), callers must use the return value.
//
// In both, source values take precedence over dest values, mirroring the Ruby
// call deep_merge!(source, dest): source is merged into dest.
//
// # Merge rules
//
// Two hashes are merged key by key. For a key present in both, the values are
// merged recursively. A key whose dest value is nil or false is treated as
// absent (Ruby truthiness), so source creates/overwrites it.
//
// Two arrays are combined. By default the result is the set union of dest then
// source (dest | source in Ruby: duplicates removed, dest order first). See the
// [Options] fields for element-wise hash merging, duplicate keeping, sorting,
// overwrite, unpacking and knockout behaviour.
//
// A type mismatch (for example a scalar meeting a hash) is unmergeable: source
// overwrites dest when Overwrite is in effect, otherwise dest is preserved.
//
// # Options
//
// Options mirrors the deep_merge gem. Overwrite and PreserveUnmergeables map to
// the gem's bang / non-bang variants: the effective "overwrite unmergeables"
// flag is Overwrite || !PreserveUnmergeables, which is true by default (like
// deep_merge!). Set PreserveUnmergeables to true (and leave Overwrite false) for
// the non-destructive-of-conflicts deep_merge behaviour.
//
// KnockoutPrefix enables the gem's knockout feature: a source string element
// prefixed with it removes the matching element from dest. Because Go cannot
// distinguish an unset string from an empty one, an empty KnockoutPrefix means
// "no knockout" rather than the gem's InvalidParameter error. Specifying a
// KnockoutPrefix while overwrite is disabled panics with [*InvalidOptionError],
// matching the gem's refusal to combine knockout with preserve_unmergeables.
//
// The package has no dependency on any Ruby runtime and imports only the
// standard library. A consumer such as go-embedded-ruby (rbgo) can adapt a Ruby
// Hash#deep_merge call onto these functions by marshalling Ruby values to the
// Go value model above.
package deepmerge
