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
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func linuxize(args []string, native bool) ([]string, error) {
	if native {
		return args, nil
	}

	return windowsArgs()
}

func windowsArgs() ([]string, error) {
	cmd := wmiGetCommandLineCmd(os.Getpid())

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	res := strings.Split(string(out), "\r\n")

	if len(res) == 0 || !strings.HasPrefix(res[0], "CommandLine") || len(res[1]) == 0 {
		return nil, errors.New("could not retrieve windows command line using wmic")
	}

	return splitToArgs(res[1]), nil
}

const (
	noRune      = rune(-1)
	space       = ' '
	quoteDouble = '"'
	quoteSingle = '\''
)

func splitToArgs(commandLine string) []string {
	args := []string{}

	currentRune := noRune
	arg := ""

	// remove excess, and not needed chars
	commandLine = strings.ReplaceAll(commandLine, "\r", "")
	commandLine = strings.TrimSpace(commandLine)

	// split
	for _, rune := range commandLine {
		if currentRune == noRune {
			switch rune {
			case space:
				if len(arg) > 0 {
					args = append(args, trimQuotes(arg))
					arg = ""
				}
			case quoteDouble:
				currentRune = quoteDouble
				arg += string(rune)
			case quoteSingle:
				currentRune = quoteSingle
				arg += string(rune)
			default:
				arg += string(rune)
			}
		} else {
			if rune == currentRune {
				currentRune = noRune
			}
			arg += string(rune)
		}
	}
	if len(arg) > 0 {
		args = append(args, trimQuotes(arg))
	}

	return args
}

func trimQuotes(s string) string {
	if strings.HasPrefix(s, string(quoteDouble)) && strings.HasSuffix(s, string(quoteDouble)) {
		s = trim(s, quoteDouble)
		s = unescape(s, quoteDouble)
	} else if strings.HasPrefix(s, string(quoteSingle)) && strings.HasSuffix(s, string(quoteSingle)) {
		s = trim(s, quoteSingle)
	}

	return s
}

func trim(s string, t rune) string {
	return strings.TrimPrefix(strings.TrimSuffix(s, string(t)), string(t))
}

func unescape(s string, t rune) string {
	return strings.ReplaceAll(s, `\`+string(t), string(t))
}

func wmiGetCommandLineCmd(processID int) exec.Cmd {
	procID := strconv.Itoa(processID)

	var cmd = exec.Command("wmic", "process",
		"where", "ProcessId="+procID,
		"get", "CommandLine")

	return *cmd
}
