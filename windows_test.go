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

import (
	"os"
	"runtime"
	"strconv"
	"testing"
)

func TestLinuxize(t *testing.T) {
	sampleOsArgs := os.Args

	isWindowsPlatform := (runtime.GOOS == "windows")

	if isWindowsPlatform {
		{ // test getting native args
			res, err := linuxize(sampleOsArgs, true)
			if err != nil {
				t.Error("failed to get native args on a windows platform")
			}
			if res == nil || len(res) == 0 {
				t.Error("no args returned")
			}
			for i, v := range res {
				if v != sampleOsArgs[i] {
					t.Errorf("wrong arg %v, expected %v, got %v", i, sampleOsArgs[i], v)
				}
			}
		}
		{ // test getting args using WMI
			res, err := linuxize(sampleOsArgs, false)
			if err != nil {
				t.Error("failed to linuxize args on a windows platform using WMI")
			}
			if res == nil || len(res) == 0 {
				t.Error("no args returned")
			}
			for i, v := range res {
				if v != sampleOsArgs[i] {
					t.Errorf("wrong arg %v, expected %v, got %v", i, sampleOsArgs[i], v)
				}
			}
		}
	} else {
		{ // test getting native args
			res, err := linuxize(sampleOsArgs, true)
			if err != nil {
				t.Error("failed to get native args on a Windows platform")
			}
			if res == nil || len(res) == 0 {
				t.Error("no args returned")
			}
			for i, v := range res {
				if v != sampleOsArgs[i] {
					t.Errorf("wrong arg %v, expected %v, got %v", i, sampleOsArgs[i], v)
				}
			}
		}
		{ // test failing to get args using WMI
			_, err := linuxize(sampleOsArgs, false)
			if err == nil {
				t.Error("no error returned calling WMI on a non-Windows platform")
			}
		}
	}
}

func TestWindowsArgs(t *testing.T) {
	res, err := windowsArgs()
	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("no error getting windows cmd args on a non windows platform [which is really creepy...]")
		}
		if res != nil && len(res) > 0 {
			t.Error("some windows args returned [really, really creepy...]")
		}
	} else {
		if err != nil {
			t.Error("error getting windows cmd args using wmi")
		}
		if res == nil || len(res) == 0 {
			t.Error("no windows args returned")
		}
	}
}

type TestSplitToArgsItem struct {
	input    string
	expected []string
}

func TestSplitToArgs(t *testing.T) {
	testItems := []TestSplitToArgsItem{
		{`command`, []string{`command`}},
		{`"command"`, []string{`command`}},
		{`command --help`, []string{`command`, `--help`}},
		{`command --help` + "\r\n", []string{`command`, `--help`}},
		{`command -t test --input "alpha"`, []string{`command`, `-t`, `test`, `--input`, `alpha`}},
		{`"command" --do '{:test :noop}'`, []string{`command`, `--do`, `{:test :noop}`}},
		{`command   --extra-spaces   `, []string{`command`, `--extra-spaces`}},
		{`command   --returns` + "\r\n" + `  ` + "\r", []string{`command`, `--returns`}},
		{`command --empty-arg ""`, []string{`command`, `--empty-arg`, ``}},
	}

	for _, v := range testItems {
		res := splitToArgs(v.input)

		if len(res) != len(v.expected) {
			t.Errorf("failed splitting to correct number of args, expected %v, got %v", len(v.expected), len(res))
		}

		for i, a := range res {
			if a != v.expected[i] {
				t.Errorf("split to args failed, arg %v, expected %v, got %v", i, v.expected[i], a)
			}
		}
	}
}

type TestTrimQuotesItem struct {
	input    string
	expected string
}

func TestTrimQuotes(t *testing.T) {
	testItems := []TestTrimQuotesItem{
		{`"double-quotes"`, `double-quotes`},
		{`'single-quotes'`, `single-quotes`},
		{`"one-double-quote`, `"one-double-quote`},
		{`'one-single-quote`, `'one-single-quote`},
		{`end-in-double-quote"`, `end-in-double-quote"`},
		{`end-in-single-quote'`, `end-in-single-quote'`},
		{`no-quotes`, `no-quotes`},
		{`in "having" quotes`, `in "having" quotes`},
		{`in 'having' quotes`, `in 'having' quotes`},
		{`"escaped \"double\" quotes"`, `escaped "double" quotes`},
	}

	for _, v := range testItems {
		res := trimQuotes(v.input)

		if res != v.expected {
			t.Errorf("quotes trim of %v failed, expected %v, got %v", v.input, v.expected, res)
		}
	}
}

type TestTrimItem struct {
	input    string
	char     rune
	expected string
}

func TestTrim(t *testing.T) {
	testItems := []TestTrimItem{
		{"-word-", '-', "word"},
		{"-word", '-', "word"},
		{"word-", '-', "word"},
		{"-two-words-", '-', "two-words"},
		{"--two-words--", '-', "-two-words-"},
		{"-three-full-words-", '$', "-three-full-words-"},
	}

	for _, v := range testItems {
		res := trim(v.input, v.char)

		if res != v.expected {
			t.Errorf("trim of %v failed, expected %v, got %v", v.input, v.expected, res)
		}
	}
}

type TestWmiArgsItem struct {
	pos      int
	expected string
}

func TestWmiGetCommandLineCmd(t *testing.T) {
	procID := 313

	cmd := wmiGetCommandLineCmd(procID)

	if len(cmd.Args) != 6 {
		t.Errorf("wrong number of wmi command arguments, expected %v, got %v", 6, len(cmd.Args))
	}

	testItems := []TestWmiArgsItem{
		{1, "process"},
		{2, "where"},
		{3, "ProcessId=" + strconv.Itoa(procID)},
		{4, "get"},
		{5, "CommandLine"},
	}

	for _, v := range testItems {
		if cmd.Args[v.pos] != v.expected {
			t.Errorf("wrong wmi command argument, expected %v, got %v", v.expected, cmd.Args[v.pos])
		}
	}
}
