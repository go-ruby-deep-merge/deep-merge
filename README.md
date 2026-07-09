# deep-merge

[![ci](https://github.com/go-ruby-deep-merge/deep-merge/actions/workflows/ci.yml/badge.svg)](https://github.com/go-ruby-deep-merge/deep-merge/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-ruby-deep-merge/deep-merge.svg)](https://pkg.go.dev/github.com/go-ruby-deep-merge/deep-merge)
[![License: BSD-3-Clause](https://img.shields.io/badge/License-BSD--3--Clause-blue.svg)](LICENSE)

A pure-Go (`CGO_ENABLED=0`, standard library only) port of the Ruby
[`deep_merge`](https://github.com/danielsdeleo/deep_merge) gem: recursive
hash/array merge with the gem's full option set.

It operates on Go values that model Ruby data:

| Ruby   | Go              |
| ------ | --------------- |
| Hash   | `map[string]any`|
| Array  | `[]any`         |
| scalar | any other value |

## Install

```sh
go get github.com/go-ruby-deep-merge/deep-merge
```

```go
import deepmerge "github.com/go-ruby-deep-merge/deep-merge"
```

## Usage

```go
dest := map[string]any{"property": []any{"2", "4"}}
source := map[string]any{"property": []any{"1", "3"}}

merged := deepmerge.DeepMerge(dest, source, deepmerge.Options{})
// merged == {"property": ["2", "4", "1", "3"]}  (dest ∪ source; dest untouched)
```

`source` is merged into `dest` (equivalent to Ruby `deep_merge!(source, dest)`):
source values take precedence.

- `DeepMerge(dest, source, opts)` — merges into a deep copy of `dest`, returns
  the copy; the caller's `dest` is left untouched.
- `DeepMergeInto(dest, source, opts)` — merges into `dest` in place and returns
  the result. Always use the return value: the top-level value can be replaced
  (for example when `dest` is `nil` or a scalar).

## Options

| Field | deep_merge option | Effect |
| ----- | ----------------- | ------ |
| `Overwrite` / `PreserveUnmergeables` | bang vs non-bang | Effective overwrite = `Overwrite \|\| !PreserveUnmergeables` (default: overwrite, like `deep_merge!`). Preserve keeps existing dest values on a type mismatch. |
| `KnockoutPrefix` | `:knockout_prefix` | A source string prefixed with it removes the matching element from dest. |
| `OverwriteArrays` | `:overwrite_arrays` | Replace dest arrays instead of merging. |
| `SortMergedArrays` | `:sort_merged_arrays` | Sort every merged array. |
| `UnpackArrays` | `:unpack_arrays` | Join + split arrays on this separator before merging. |
| `MergeHashArrays` | `:merge_hash_arrays` | Merge arrays of hashes element-wise (by index). |
| `ExtendExistingArrays` | `:extend_existing_arrays` | Push a source hash/scalar onto a dest array. |
| `KeepArrayDuplicates` | `:keep_array_duplicates` | Concatenate arrays keeping duplicates instead of taking the union. |
| `MergeNilValues` | `:merge_nil_values` | Merge `nil` source values instead of skipping them. |

### Notes on fidelity

- Ruby truthiness is honoured: a dest key whose value is `nil` or `false` is
  treated as absent, so source creates/overwrites it.
- Because Go cannot distinguish an unset string from an empty one, an empty
  `KnockoutPrefix` means "no knockout" rather than the gem's `InvalidParameter`
  error. Combining a `KnockoutPrefix` with disabled overwrite panics with
  `*InvalidOptionError`, matching the gem's refusal of that combination.
- Ruby Hashes with non-string keys (e.g. symbols, integers) are out of scope;
  keys are `string`.

## License

BSD-3-Clause. See [LICENSE](LICENSE).
