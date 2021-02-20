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
	"fmt"
	"os"
)

func init() {
	var err error
	tools4CljDir, err = getTools4CljPath()
	if err != nil {
		panic(err)
	}

	javaPath, err = getJavaPath()
	if err != nil {
		panic(err)
	}

	toolsCp, err = getToolsCp(tools4CljDir)
	if err != nil {
		panic(err)
	}

	execCp, err = getExecCp(tools4CljDir)
	if err != nil {
		panic(err)
	}
}

// RunClojure runs clojure using the official clojure tools
func RunClojure(osArgs []string) {
	runClojure(osArgs, false)
}

// RunClj runs clojure within a readline wrapper
func RunClj(osArgs []string) {
	runClojure(osArgs, true)
}

func runClojure(osArgs []string, cljRun bool) {
	// download official clojure tools
	err := getClojureTools(tools4CljDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var opts allOpts

	// read and set command line options
	exit, err := read(&opts, osArgs, cljRun)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if exit {
		os.Exit(0)
	}

	if opts.Main.Help {
		fmt.Println(usage)
		return
	}

	// use command line options
	err = use(&opts)
	if err != nil {
		if opts.Clj.Verbose {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
