// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-deep-merge/deep-merge authors

package deepmerge

import (
	"reflect"
	"testing"
)

// Concise aliases for the Ruby value model.
type m = map[string]any
type a = []any

// merge case table. Each case is the Go equivalent of a Ruby
// DeepMerge::deep_merge!(source, dest, opts) call, i.e. DeepMergeInto(dest,
// source, opts), asserted against want. Cases mirror danielsdeleo/deep_merge's
// test/test_deep_merge.rb (string keys used throughout).
type mergeCase struct {
	name   string
	dest   any
	source any
	opts   Options
	want   any
}

var koPrefix = "--"

func mergeCases() []mergeCase {
	return []mergeCase{
		// --- basics ---------------------------------------------------------
		{"knockout merge ints", m{"id": a{1, 2, 3}}, m{"id": a{3, 4, 5}},
			Options{KnockoutPrefix: koPrefix}, m{"id": a{1, 2, 3, 4, 5}}},
		{"bang merge ints", m{"id": a{1, 2, 3}}, m{"id": a{3, 4, 5}},
			Options{}, m{"id": a{1, 2, 3, 4, 5}}},
		{"preserve string vs array", m{"id": a{1, 2, 3}}, m{"id": "xxx"},
			Options{PreserveUnmergeables: true}, m{"id": a{1, 2, 3}}},

		{"merge empty source", m{"property": a{"2", "4"}}, m{},
			Options{}, m{"property": a{"2", "4"}}},
		{"merge into empty dest", m{}, m{"property": a{"2", "4"}},
			Options{}, m{"property": a{"2", "4"}}},
		{"simple string overwrite", m{"name": "value1"}, m{"name": "value"},
			Options{}, m{"name": "value"}},
		{"string into empty dest", m{}, m{"name": "value"},
			Options{}, m{"name": "value"}},

		{"arrays union", m{"property": a{"2", "4"}}, m{"property": a{"1", "3"}},
			Options{}, m{"property": a{"2", "4", "1", "3"}}},
		{"arrays overwrite", m{"property": a{"2", "4"}}, m{"property": a{"1", "3"}},
			Options{OverwriteArrays: true}, m{"property": a{"1", "3"}}},
		{"arrays sorted", m{"property": a{"2", "4"}}, m{"property": a{"1", "3"}},
			Options{SortMergedArrays: true}, m{"property": a{"1", "2", "3", "4"}}},

		// --- nested hashes / type interactions ------------------------------
		{"nested arrays union",
			m{"property": m{"bedroom_count": a{"3", "2"}, "bathroom_count": a{"2"}}},
			m{"property": m{"bedroom_count": a{"1", "2"}, "bathroom_count": a{"1", "4+"}}},
			Options{},
			m{"property": m{"bedroom_count": a{"3", "2", "1"}, "bathroom_count": a{"2", "1", "4+"}}}},
		{"array overwrites string (bang)",
			m{"property": m{"bedroom_count": "3", "bathroom_count": a{"2"}}},
			m{"property": m{"bedroom_count": a{"1", "2"}, "bathroom_count": a{"1", "4+"}}},
			Options{},
			m{"property": m{"bedroom_count": a{"1", "2"}, "bathroom_count": a{"2", "1", "4+"}}}},
		{"array does not overwrite string (preserve)",
			m{"property": m{"bedroom_count": "3", "bathroom_count": a{"2"}}},
			m{"property": m{"bedroom_count": a{"1", "2"}, "bathroom_count": a{"1", "4+"}}},
			Options{PreserveUnmergeables: true},
			m{"property": m{"bedroom_count": "3", "bathroom_count": a{"2", "1", "4+"}}}},
		{"string overwrites array (bang)",
			m{"property": m{"bedroom_count": a{"1", "2"}, "bathroom_count": a{"2"}}},
			m{"property": m{"bedroom_count": "3", "bathroom_count": a{"1", "4+"}}},
			Options{},
			m{"property": m{"bedroom_count": "3", "bathroom_count": a{"2", "1", "4+"}}}},
		{"string does not overwrite array (preserve)",
			m{"property": m{"bedroom_count": a{"1", "2"}, "bathroom_count": a{"2"}}},
			m{"property": m{"bedroom_count": "3", "bathroom_count": a{"1", "4+"}}},
			Options{PreserveUnmergeables: true},
			m{"property": m{"bedroom_count": a{"1", "2"}, "bathroom_count": a{"2", "1", "4+"}}}},
		{"hash overwrites array (bang)",
			m{"property": m{"bedroom_count": a{"1", "2"}, "bathroom_count": a{"2"}}},
			m{"property": m{"bedroom_count": m{"king_bed": 3, "queen_bed": 1}, "bathroom_count": a{"1", "4+"}}},
			Options{},
			m{"property": m{"bedroom_count": m{"king_bed": 3, "queen_bed": 1}, "bathroom_count": a{"2", "1", "4+"}}}},
		{"hash does not overwrite array (preserve)",
			m{"property": m{"bedroom_count": a{"1", "2"}, "bathroom_count": a{"2"}}},
			m{"property": m{"bedroom_count": m{"king_bed": 3, "queen_bed": 1}, "bathroom_count": a{"1", "4+"}}},
			Options{PreserveUnmergeables: true},
			m{"property": m{"bedroom_count": a{"1", "2"}, "bathroom_count": a{"2", "1", "4+"}}}},
		{"deep ints overwritten",
			m{"property": m{"bedroom_count": m{"king_bed": 2, "queen_bed": 4}, "bathroom_count": a{"2"}}},
			m{"property": m{"bedroom_count": m{"king_bed": 3, "queen_bed": 1}, "bathroom_count": a{"1", "4+"}}},
			Options{},
			m{"property": m{"bedroom_count": m{"king_bed": 3, "queen_bed": 1}, "bathroom_count": a{"2", "1", "4+"}}}},
		{"deep arrays merged",
			m{"property": m{"bedroom_count": m{"king_bed": a{2}, "queen_bed": a{4}}, "bathroom_count": a{"2"}}},
			m{"property": m{"bedroom_count": m{"king_bed": a{3}, "queen_bed": a{1}}, "bathroom_count": a{"1", "4+"}}},
			Options{},
			m{"property": m{"bedroom_count": m{"king_bed": a{2, 3}, "queen_bed": a{4, 1}}, "bathroom_count": a{"2", "1", "4+"}}}},
		{"scalar overwrites whole subtree (bang)",
			m{"property": m{"bedroom_count": m{"king_bed": a{2}}, "bathroom_count": a{"2"}}},
			m{"property": "1"}, Options{}, m{"property": "1"}},
		{"scalar does not overwrite subtree (preserve)",
			m{"property": m{"bedroom_count": m{"king_bed": a{2}}, "bathroom_count": a{"2"}}},
			m{"property": "1"}, Options{PreserveUnmergeables: true},
			m{"property": m{"bedroom_count": m{"king_bed": a{2}}, "bathroom_count": a{"2"}}}},
		{"source incomplete keeps dest keys",
			m{"property": m{"bedroom_count": m{"king_bed": a{2}, "queen_bed": a{4}}, "bathroom_count": a{"2"}}},
			m{"property": m{"bedroom_count": m{"king_bed": a{3}}, "bathroom_count": a{"1"}}},
			Options{},
			m{"property": m{"bedroom_count": m{"king_bed": a{2, 3}, "queen_bed": a{4}}, "bathroom_count": a{"2", "1"}}}},
		{"empty source deep noop",
			m{"property": m{"bedroom_count": m{"king_bed": a{2}}, "bathroom_count": a{"2"}}},
			m{}, Options{},
			m{"property": m{"bedroom_count": m{"king_bed": a{2}}, "bathroom_count": a{"2"}}}},
		{"empty dest deep copy",
			m{}, m{"property": m{"bedroom_count": m{"king_bed": a{3}}, "bathroom_count": a{"1"}}},
			Options{},
			m{"property": m{"bedroom_count": m{"king_bed": a{3}}, "bathroom_count": a{"1"}}}},

		// --- nils inside arrays --------------------------------------------
		{"nil in source array",
			m{"a": m{"kb": a{2}, "qb": a{4}}, "bc": a{"2"}},
			m{"a": m{"kb": a{nil}, "qb": a{1, nil}}, "bc": a{nil, "1"}},
			Options{},
			m{"a": m{"kb": a{2, nil}, "qb": a{4, 1, nil}}, "bc": a{"2", nil, "1"}}},
		{"nil in dest array",
			m{"a": m{"kb": a{nil}, "qb": a{4, nil}}, "bc": a{nil, "2"}},
			m{"a": m{"kb": a{3}, "qb": a{1}}, "bc": a{"1"}},
			Options{},
			m{"a": m{"kb": a{nil, 3}, "qb": a{4, nil, 1}}, "bc": a{nil, "2", "1"}}},

		// --- extend_existing_arrays ----------------------------------------
		{"extend array with scalar", m{"property": a{"1", "2", "3"}}, m{"property": "4"},
			Options{ExtendExistingArrays: true}, m{"property": a{"1", "2", "3", "4"}}},
		{"extend array with hash",
			m{"property": a{m{"number": "1"}, m{"number": "2"}}},
			m{"property": m{"number": "3"}},
			Options{ExtendExistingArrays: true},
			m{"property": a{m{"number": "1"}, m{"number": "2"}, m{"number": "3"}}}},

		// --- unpack_arrays -------------------------------------------------
		{"unpack source and dest",
			m{}, m{"property": m{"bedroom_count": a{"1", "2,3"}}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"property": m{"bedroom_count": a{"1", "2", "3"}}}},
		{"unpack disabled keeps compound",
			m{}, m{"property": m{"bedroom_count": a{"1", "2,3"}}},
			Options{KnockoutPrefix: koPrefix},
			m{"property": m{"bedroom_count": a{"1", "2,3"}}}},
		{"unpacked edge without knockout",
			m{"region": m{"ids": a{"1", "2", "3", "4"}, "id": "11"}},
			m{"region": m{"ids": a{"7", "--", "2", "6,8"}}},
			Options{UnpackArrays: ","},
			m{"region": m{"ids": a{"1", "2", "3", "4", "7", "--", "6", "8"}, "id": "11"}}},
		{"no split keeps compound edge",
			m{"region": m{"ids": a{"1", "2", "3", "4"}, "id": "11"}},
			m{"region": m{"ids": a{"7", "3", "--", "6,8"}}},
			Options{},
			m{"region": m{"ids": a{"1", "2", "3", "4", "7", "--", "6,8"}, "id": "11"}}},

		// --- knockout: arrays ----------------------------------------------
		{"knockout removes matching",
			m{"property": m{"bedroom_count": a{"1", "2", "3"}}},
			m{"property": m{"bedroom_count": a{"--1", "2", "3"}}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"property": m{"bedroom_count": a{"2", "3"}}}},
		{"knockout absent then add",
			m{"property": m{"bedroom_count": a{"4"}}},
			m{"property": m{"bedroom_count": a{"--1", "2", "3"}}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"property": m{"bedroom_count": a{"4", "2", "3"}}}},
		{"knockout both ids",
			m{"amenity": m{"id": a{"1", "2"}}},
			m{"amenity": m{"id": a{"--1", "--2", "3", "4"}}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"amenity": m{"id": a{"3", "4"}}}},
		{"knockout compound unpack",
			m{"amenity": m{"id": a{"1", "2"}}},
			m{"amenity": m{"id": a{"--1,--2", "3,4"}}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"amenity": m{"id": a{"3", "4"}}}},
		{"knockout compound without unpack",
			m{"amenity": m{"id": a{"1", "2"}}},
			m{"amenity": m{"id": a{"--1,--2", "3,4"}}},
			Options{KnockoutPrefix: koPrefix},
			m{"amenity": m{"id": a{"1", "2", "3,4"}}}},
		{"knockout mixed tokens",
			m{"amenity": m{"id": a{"1", "2"}}},
			m{"amenity": m{"id": a{"--1,2", "3,4", "--5", "6"}}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"amenity": m{"id": a{"2", "3", "4", "6"}}}},

		// --- knockout: naked token truncates dest --------------------------
		{"naked knockout string on string",
			m{"amenity": "1"}, m{"amenity": "--"},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"amenity": ""}},
		{"naked knockout array on string",
			m{"amenity": "1"}, m{"amenity": a{"--"}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"amenity": a{}}},
		{"naked knockout string on array",
			m{"amenity": a{"1"}}, m{"amenity": "--"},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"amenity": ""}},
		{"naked knockout array on array",
			m{"amenity": a{"1"}}, m{"amenity": a{"--"}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"amenity": a{}}},
		{"naked knockout then keep two",
			m{"amenity": a{"1", "3", "7+"}}, m{"amenity": a{"--", "2"}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"amenity": a{"2"}}},
		{"naked knockout on scalar dest keeps source",
			m{"amenity": "5"}, m{"amenity": a{"--", "2"}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"amenity": a{"2"}}},
		{"naked knockout string on hash",
			m{"amenity": m{"id": a{"1", "2"}}}, m{"amenity": "--"},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"amenity": ""}},
		{"naked knockout array on hash",
			m{"amenity": m{"id": a{"1", "2"}}}, m{"amenity": a{"--"}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"amenity": a{}}},
		{"knockout dest array to empty string",
			m{"region": m{"ids": a{"1", "2", "3", "4"}}}, m{"region": m{"ids": "--"}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"region": m{"ids": ""}}},
		{"knockout array leaves siblings",
			m{"region": m{"ids": a{"1", "2", "3", "4"}, "id": "11"}}, m{"region": m{"ids": "--"}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"region": m{"ids": "", "id": "11"}}},
		{"knockout whole subtree",
			m{"region": m{"ids": a{"1", "2"}, "id": "11"}}, m{"region": "--"},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"region": ""}},
		{"knockout subtree retain array form",
			m{"region": m{"ids": a{"1", "2", "3", "4"}, "id": "11"}}, m{"region": m{"ids": a{"--"}}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"region": m{"ids": a{}, "id": "11"}}},
		{"knockout then replace content",
			m{"region": m{"ids": a{"1", "2", "3", "4"}, "id": "11"}}, m{"region": m{"ids": a{"2", "--", "6"}}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"region": m{"ids": a{"2", "6"}, "id": "11"}}},

		// --- knockout: scalar strings --------------------------------------
		{"knockout scalar match",
			m{"amenity": "1"}, m{"amenity": "--1"},
			Options{KnockoutPrefix: koPrefix}, m{"amenity": ""}},
		{"knockout scalar mismatch",
			m{"amenity": "2"}, m{"amenity": "--1"},
			Options{KnockoutPrefix: koPrefix}, m{"amenity": ""}},
		{"knockout scalar into empty",
			m{}, m{"amenity": "--1"},
			Options{KnockoutPrefix: koPrefix}, m{"amenity": ""}},

		// --- knockout: real-world session merges ---------------------------
		{"overwrite unmergeables into empty",
			m{}, m{"action": "browse", "controller": "results"},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"action": "browse", "controller": "results"}},
		{"session merge with url_regions",
			m{"region": m{"ids": a{"227"}}},
			m{"url_regions": a{}, "region": m{"ids": a{"227,233"}}, "action": "browse"},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"url_regions": a{}, "region": m{"ids": a{"227", "233"}}, "action": "browse"}},
		{"knockout dedup ids",
			m{"region": m{"ids": a{"227", "233", "324", "230", "230"}, "id": "230"}},
			m{"region": m{"ids": a{"--", "227"}, "id": "230"}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"region": m{"ids": a{"227"}, "id": "230"}}},
		{"knockout empty ids preserve id",
			m{"region": m{"muni_city_id": "2244", "ids": a{"227", "2", "3", "3"}, "id": "3"}, "query_uuid": "zzz"},
			m{"region": m{"ids": a{"--"}}, "query_uuid": "zzz"},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"region": m{"muni_city_id": "2244", "ids": a{}, "id": "3"}, "query_uuid": "zzz"}},
		{"knockout string ids and city",
			m{"region": m{"muni_city_id": "2244", "ids": a{"227", "2", "3", "3"}, "id": "3"}, "query_uuid": "zzz"},
			m{"region": m{"muni_city_id": "--", "ids": "--", "id": "5"}, "query_uuid": "zzz"},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"region": m{"muni_city_id": "", "ids": "", "id": "5"}, "query_uuid": "zzz"}},
		{"empty strings unaffected by knockout",
			m{"muni_city_id": "", "id": ""}, m{"muni_city_id": "--", "id": ""},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"muni_city_id": "", "id": ""}},
		{"knockout suppresses prefixed values on new key",
			m{}, m{"amenity": m{"id": a{"--26,--27,28"}}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","},
			m{"amenity": m{"id": a{"28"}}}},
		{"unmerged hash without unpack",
			m{}, m{"amenity": m{"id": a{"26,27"}}},
			Options{KnockoutPrefix: koPrefix}, m{"amenity": m{"id": a{"26,27"}}}},
		{"overwrite unmergeables array residual knockout",
			m{"amenity": "5"}, m{"amenity": a{"--", "--x"}},
			Options{KnockoutPrefix: koPrefix, UnpackArrays: ","}, m{"amenity": a{}}},

		// --- merge_hash_arrays ---------------------------------------------
		{"array of hashes appended (no option)",
			m{"item": a{m{"3": "5"}}}, m{"item": a{m{"1": "3"}, m{"2": "4"}}},
			Options{}, m{"item": a{m{"3": "5"}, m{"1": "3"}, m{"2": "4"}}}},
		{"merge hash arrays single",
			m{"item": a{m{"3": "5"}}}, m{"item": a{m{"1": "3"}}},
			Options{MergeHashArrays: true}, m{"item": a{m{"3": "5", "1": "3"}}}},
		{"merge hash arrays source longer",
			m{"item": a{m{"3": "5"}}}, m{"item": a{m{"1": "3"}, m{"2": "4"}}},
			Options{MergeHashArrays: true}, m{"item": a{m{"3": "5", "1": "3"}, m{"2": "4"}}}},
		{"merge hash arrays dest longer",
			m{"item": a{m{"3": "5"}, m{"2": "4"}}}, m{"item": a{m{"1": "3"}}},
			Options{MergeHashArrays: true}, m{"item": a{m{"3": "5", "1": "3"}, m{"2": "4"}}}},
		{"merge hash arrays with non-hash falls back",
			m{"item": a{m{"3": "5"}}}, m{"item": a{m{"1": "3"}, "str"}},
			Options{MergeHashArrays: true}, m{"item": a{m{"3": "5"}, m{"1": "3"}, "str"}}},

		// --- keep_array_duplicates -----------------------------------------
		{"keep duplicates concat",
			m{"item": a{"1", "2"}}, m{"item": a{"2", "3"}},
			Options{KeepArrayDuplicates: true}, m{"item": a{"1", "2", "2", "3"}}},
		{"keep duplicates new key",
			m{}, m{"item": a{"2", "3"}},
			Options{KeepArrayDuplicates: true}, m{"item": a{"2", "3"}}},

		// --- nil handling ---------------------------------------------------
		{"nil skipped by default",
			m{"item": "existing"}, m{"item": nil}, Options{}, m{"item": "existing"}},
		{"nil merged when enabled",
			m{"item": "existing"}, m{"item": nil},
			Options{MergeNilValues: true}, m{"item": nil}},

		// --- branch coverage extras ----------------------------------------
		{"empty string overwrites (bang)",
			m{"item": "hello"}, m{"item": ""}, Options{}, m{"item": ""}},
		{"array into scalar preserve keeps dest",
			m{"x": "s"}, m{"x": a{1, 2}}, Options{PreserveUnmergeables: true}, m{"x": "s"}},
		{"hash into scalar overwrite",
			m{"x": "s"}, m{"x": m{"y": 1}}, Options{}, m{"x": m{"y": 1}}},
		{"hash into scalar overwrite with knockout",
			m{"x": "s"}, m{"x": m{"y": 1}}, Options{KnockoutPrefix: koPrefix}, m{"x": m{"y": 1}}},
		{"int into scalar overwrite with knockout",
			m{"x": "s"}, m{"x": 7}, Options{KnockoutPrefix: koPrefix}, m{"x": 7}},
		{"hash into scalar preserve keeps dest",
			m{"x": "s"}, m{"x": m{"y": 1}}, Options{PreserveUnmergeables: true}, m{"x": "s"}},
		{"scalar into scalar preserve keeps dest",
			m{"x": "a"}, m{"x": "b"}, Options{PreserveUnmergeables: true}, m{"x": "a"}},
		{"false dest treated as absent",
			m{"x": false}, m{"x": 1}, Options{}, m{"x": 1}},
		{"nil dest value replaced",
			m{"x": nil}, m{"x": 1}, Options{}, m{"x": 1}},
		{"new-key non-keep dedups source array",
			m{}, m{"x": a{"1", "1", "2"}}, Options{}, m{"x": a{"1", "2"}}},
	}
}

func TestDeepMergeInto(t *testing.T) {
	for _, tc := range mergeCases() {
		t.Run(tc.name, func(t *testing.T) {
			got := DeepMergeInto(tc.dest, tc.source, tc.opts)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestDeepMergeNonDestructive(t *testing.T) {
	dest := m{"a": a{"1", "2"}, "b": m{"c": "keep"}}
	source := m{"a": a{"3"}, "b": m{"d": "add"}}
	got := DeepMerge(dest, source, Options{})
	want := m{"a": a{"1", "2", "3"}, "b": m{"c": "keep", "d": "add"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	// Original dest must be untouched.
	orig := m{"a": a{"1", "2"}, "b": m{"c": "keep"}}
	if !reflect.DeepEqual(dest, orig) {
		t.Fatalf("dest mutated: %#v", dest)
	}
}

func TestDeepMergeIntoTopLevelReplacements(t *testing.T) {
	if got := DeepMergeInto(nil, m{"x": 1}, Options{}); !reflect.DeepEqual(got, m{"x": 1}) {
		t.Fatalf("nil dest: got %#v", got)
	}
	if got := DeepMergeInto("scalar", m{"x": 1}, Options{}); !reflect.DeepEqual(got, m{"x": 1}) {
		t.Fatalf("scalar dest overwrite: got %#v", got)
	}
	if got := DeepMergeInto("scalar", m{"x": 1}, Options{PreserveUnmergeables: true}); got != "scalar" {
		t.Fatalf("scalar dest preserve: got %#v", got)
	}
	// Array dest, array source at top level.
	if got := DeepMergeInto(a{"1"}, a{"2"}, Options{}); !reflect.DeepEqual(got, a{"1", "2"}) {
		t.Fatalf("array top level: got %#v", got)
	}
}

func TestInvalidOptionPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		err, ok := r.(*InvalidOptionError)
		if !ok {
			t.Fatalf("expected *InvalidOptionError, got %T", r)
		}
		if err.Error() == "" {
			t.Fatal("empty error message")
		}
	}()
	DeepMergeInto(m{}, m{}, Options{KnockoutPrefix: "--", PreserveUnmergeables: true})
}

func TestOverwriteDisablesPreserve(t *testing.T) {
	// Overwrite=true forces overwrite even with PreserveUnmergeables set, so
	// knockout is permitted and conflicting scalars are overwritten.
	got := DeepMergeInto(m{"x": "old"}, m{"x": "new"},
		Options{Overwrite: true, PreserveUnmergeables: true})
	if !reflect.DeepEqual(got, m{"x": "new"}) {
		t.Fatalf("got %#v", got)
	}
}

// --- white-box helper tests for exhaustive branch coverage -----------------

func TestTruthy(t *testing.T) {
	cases := []struct {
		v    any
		want bool
	}{{nil, false}, {false, false}, {true, true}, {0, true}, {"", true}, {a{}, true}}
	for _, c := range cases {
		if got := truthy(c.v); got != c.want {
			t.Fatalf("truthy(%#v)=%v want %v", c.v, got, c.want)
		}
	}
}

func TestRubySplit(t *testing.T) {
	cases := []struct {
		s, sep string
		want   []any
	}{
		{"", ",", a{}},
		{"a,b,c", ",", a{"a", "b", "c"}},
		{"a,", ",", a{"a"}},            // trailing empty dropped
		{",", ",", a{}},                // both empty dropped
		{"1,,3", ",", a{"1", "", "3"}}, // interior empty kept
	}
	for _, c := range cases {
		if got := rubySplit(c.s, c.sep); !reflect.DeepEqual(got, c.want) {
			t.Fatalf("rubySplit(%q)=%#v want %#v", c.s, got, c.want)
		}
	}
}

func TestUnpackArrayTypes(t *testing.T) {
	got := unpackArray(a{1, "b", nil}, ",")
	want := a{"1", "b"} // 1->"1", "b", nil->"" ; trailing empty dropped
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestToRubyString(t *testing.T) {
	if toRubyString(nil) != "" {
		t.Fatal("nil")
	}
	if toRubyString("x") != "x" {
		t.Fatal("string")
	}
	if toRubyString(42) != "42" {
		t.Fatal("int")
	}
}

func TestCompareValuesAndSort(t *testing.T) {
	// numeric (int + float64), strings, and mixed fallback.
	nums := a{3, 1, 2}
	sortArray(nums)
	if !reflect.DeepEqual(nums, a{1, 2, 3}) {
		t.Fatalf("num sort %#v", nums)
	}
	floats := a{2.5, 1.5, 2.5}
	sortArray(floats)
	if !reflect.DeepEqual(floats, a{1.5, 2.5, 2.5}) {
		t.Fatalf("float sort %#v", floats)
	}
	strs := a{"c", "a", "b"}
	sortArray(strs)
	if !reflect.DeepEqual(strs, a{"a", "b", "c"}) {
		t.Fatalf("str sort %#v", strs)
	}
	// equal comparison returns 0
	if compareValues(2, 2) != 0 {
		t.Fatal("equal numbers")
	}
	if compareValues(2, 5) != -1 || compareValues(5, 2) != 1 {
		t.Fatal("numeric order")
	}
	// numeric vs string -> fallback string compare
	if compareValues(1, "1") != 0 {
		t.Fatalf("mixed fallback: %d", compareValues(1, "1"))
	}
	// two non-numeric non-string values -> fallback
	if compareValues(true, false) == 0 {
		t.Fatal("bool fallback should differ")
	}
}

func TestAsFloat(t *testing.T) {
	if v, ok := asFloat(3); !ok || v != 3 {
		t.Fatal("int")
	}
	if v, ok := asFloat(2.5); !ok || v != 2.5 {
		t.Fatal("float64")
	}
	if _, ok := asFloat("x"); ok {
		t.Fatal("string should not be float")
	}
}

func TestClearOrNil(t *testing.T) {
	if got := clearOrNil(a{1}); !reflect.DeepEqual(got, a{}) {
		t.Fatalf("array %#v", got)
	}
	if got := clearOrNil(m{"a": 1}); !reflect.DeepEqual(got, m{}) {
		t.Fatalf("map %#v", got)
	}
	if got := clearOrNil("x"); got != "" {
		t.Fatalf("string %#v", got)
	}
	if got := clearOrNil(5); got != nil {
		t.Fatalf("int %#v", got)
	}
}

func TestDeepCopyIndependence(t *testing.T) {
	src := m{"a": a{m{"b": 1}}}
	cp := deepCopy(src).(m)
	cp["a"].(a)[0].(m)["b"] = 99
	if src["a"].(a)[0].(m)["b"] != 1 {
		t.Fatal("deepCopy shared nested state")
	}
	// scalar passthrough
	if deepCopy(7) != 7 {
		t.Fatal("scalar copy")
	}
}

func TestAllHashes(t *testing.T) {
	if !allHashes(a{m{}, m{"x": 1}}) {
		t.Fatal("all hashes")
	}
	if allHashes(a{m{}, "x"}) {
		t.Fatal("not all hashes")
	}
	if !allHashes(a{}) {
		t.Fatal("empty is vacuously all hashes")
	}
}

func TestArraySetOps(t *testing.T) {
	if got := union(a{"a", "b"}, a{"b", "c"}); !reflect.DeepEqual(got, a{"a", "b", "c"}) {
		t.Fatalf("union %#v", got)
	}
	if got := concat(a{"a"}, a{"a", "b"}); !reflect.DeepEqual(got, a{"a", "a", "b"}) {
		t.Fatalf("concat %#v", got)
	}
	if !containsValue(a{1, 2}, 2) || containsValue(a{1, 2}, 3) {
		t.Fatal("containsValue")
	}
	if got := removeValue(a{1, 2, 1}, 1); !reflect.DeepEqual(got, a{2}) {
		t.Fatalf("removeValue %#v", got)
	}
}

func TestPush(t *testing.T) {
	base := a{"1"}
	got := push(base, m{"n": 1})
	if !reflect.DeepEqual(got, a{"1", m{"n": 1}}) {
		t.Fatalf("push %#v", got)
	}
	if len(base) != 1 {
		t.Fatal("push mutated base")
	}
}

func TestKnockoutHelper(t *testing.T) {
	src, dst := knockout(a{"--1", "2", 3}, a{"1", "2"}, "--")
	if !reflect.DeepEqual(src, a{"2", 3}) {
		t.Fatalf("filtered source %#v", src)
	}
	if !reflect.DeepEqual(dst, a{"2"}) {
		t.Fatalf("filtered dest %#v", dst)
	}
}

func TestRemoveKnockoutElems(t *testing.T) {
	got := removeKnockoutElems(a{"--x", "y", 3}, "--")
	if !reflect.DeepEqual(got, a{"y", 3}) {
		t.Fatalf("got %#v", got)
	}
}

func TestMergeScalarIntoNilPreserve(t *testing.T) {
	// merge default branch reaching overwriteUnmergeables with overwrite off.
	if got := merge("x", nil, Options{PreserveUnmergeables: true}.resolve()); got != nil {
		t.Fatalf("got %#v", got)
	}
}
