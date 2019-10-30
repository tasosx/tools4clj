/*************************************************************************
 * Copyright (c) 2019 Tasos Mamaloukos.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution.
 *
 * The Eclipse Public License is available at
 *     https://www.eclipse.org/org/documents/epl-v10.html
 *
 *************************************************************************/

package tools4clj

import "testing"

type TestJoinItem struct {
	input    []string
	char     string
	expected string
}

func TestJoin(t *testing.T) {
	testItems := []TestJoinItem{
		{[]string{"alpha"}, "-", "alpha"},
		{[]string{"alpha", "theta", "delta"}, "/", "alpha/theta/delta"},
		{[]string{"alpha", "theta"}, "/in the middle/", "alpha/in the middle/theta"},
		{[]string{"No", "Separator"}, "", "NoSeparator"},
	}

	for _, v := range testItems {
		res := join(v.input, v.char)
		if res != v.expected {
			t.Errorf("join failed, expected %v, got %v", v.expected, res)
		}
	}
}

type TestRemoveEmptySlicesItem struct {
	input    []string
	expected []string
}

func TestRemoveEmpty(t *testing.T) {
	testItems := []TestRemoveEmptySlicesItem{
		{[]string{"not", "empty"}, []string{"not", "empty"}},
		{[]string{"one", "", "empty"}, []string{"one", "empty"}},
		{[]string{"last", "empty", ""}, []string{"last", "empty"}},
		{[]string{"", "", "", ""}, []string{}},
		{[]string{"", "two", "words", ""}, []string{"two", "words"}},
		{[]string{}, []string{}},
	}

	for _, v := range testItems {
		res := removeEmpty(v.input)
		if len(res) != len(v.expected) {
			t.Errorf("remove empty slices failed, expected %v, got %v", len(v.expected), len(res))
		}
		for i, s := range res {
			if s != v.expected[i] {
				t.Errorf("remove empty slices failed, expected slice %v, got %v", v.expected[i], s)
			}
		}
	}
}
