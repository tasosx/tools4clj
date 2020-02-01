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
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type allOpts struct {
	Dep        depOpts
	Init       initOpts
	Main       mainOpts
	Args       []string
	NativeArgs bool
	Rlwrap     bool
}

type depOpts struct {
	JvmOpts          []string
	JvmAliases       string
	ResolveAliases   string
	ClassPathAliases string
	MainAliases      string
	AllAliases       string
	DepsData         string
	PrintClassPath   bool
	ForceCP          string
	Repro            bool
	Force            bool
	Pom              bool
	Tree             bool
	ResolveTags      bool
	Verbose          bool
	Describe         bool
	Threads          int
	Trace            bool
	InvalidOption    string
}

type initOpts struct {
	Init   string
	Eval   string
	Report string
}

type mainOpts struct {
	MainArgs []string
	Repl     bool
	Help     bool
	HelpArg  string
	PathArg  string
}

func read(all *allOpts, args []string, cljRun bool) error {
	if len(args) == 0 {
		return errors.New("missing application argument (0)")
	}

	var i = 1

	i, err := setT4COpts(all, args, i, cljRun)
	if err != nil {
		return err
	}

	// resolve "linuxized" windows command line args
	args, err = linuxize(args, all.NativeArgs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	i, err = setDepOpts(all, args, i)
	if err != nil {
		return err
	}

	i, err = setInitOpts(all, args, i)
	if err != nil {
		return err
	}

	i, err = setMainOpts(all, args, i)
	if err != nil {
		return err
	}

	if all.Main.Help && (len(all.Dep.MainAliases) > 0 || len(all.Dep.AllAliases) > 0) {
		all.Main.Help = false
		all.Args = append(all.Args, all.Main.HelpArg)
		return nil
	}

	if i < len(args) {
		all.Args = append(all.Args, args[i:]...)
	}

	return nil
}

func setT4COpts(all *allOpts, args []string, pos int, cljRun bool) (int, error) {
	all.Rlwrap = cljRun
	all.NativeArgs = (runtime.GOOS != "windows")

	rebel := false
	for {
		if pos >= len(args) {
			break
		}

		if args[pos] == "--rebel" {
			if cljRun == false {
				return pos, errors.New("readline option " + args[pos] + " can only be used with clj")
			}
			if rebel == true {
				return pos, errors.New("readline option " + args[pos] + " defined more than one time")
			}

			all.Rlwrap = false
			all.Dep.DepsData = rebelSdepsArg
			all.Main.MainArgs = append(all.Main.MainArgs, "-m", rebelMainArg)

			rebel = true

		} else if args[pos] == "--native-args" {
			all.NativeArgs = true
		} else {
			// move to the next options group
			break
		}

		pos++
	}
	return pos, nil
}

func setDepOpts(all *allOpts, args []string, pos int) (int, error) {
	for {
		if pos >= len(args) {
			break
		}

		if strings.HasPrefix(args[pos], "-J") {
			all.Dep.JvmOpts = append(all.Dep.JvmOpts, strings.TrimPrefix(args[pos], "-J"))
		} else if strings.HasPrefix(args[pos], "-O") {
			all.Dep.JvmAliases += strings.TrimPrefix(args[pos], "-O")
		} else if strings.HasPrefix(args[pos], "-R") {
			all.Dep.ResolveAliases += strings.TrimPrefix(args[pos], "-R")
		} else if strings.HasPrefix(args[pos], "-C") {
			all.Dep.ClassPathAliases += strings.TrimPrefix(args[pos], "-C")
		} else if strings.HasPrefix(args[pos], "-M") {
			all.Dep.MainAliases += strings.TrimPrefix(args[pos], "-M")
		} else if strings.HasPrefix(args[pos], "-A") {
			all.Dep.AllAliases += strings.TrimPrefix(args[pos], "-A")
		} else if args[pos] == "-Sdeps" {
			if len(all.Dep.DepsData) > 0 {
				return pos, errors.New("deps data option " + args[pos] + " defined more than one time")
			}
			if pos+1 > len(args)-1 {
				return pos, errors.New("deps data value (EDN) not defined for -Sdeps option")
			}
			pos++
			all.Dep.DepsData = args[pos]
		} else if args[pos] == "-Spath" {
			all.Dep.PrintClassPath = true
		} else if args[pos] == "-Scp" {
			if len(all.Dep.ForceCP) > 0 {
				return pos, errors.New("classpath option " + args[pos] + " defined more than one time")
			}
			if pos+1 > len(args)-1 {
				return pos, errors.New("classpath value (CP) not defined for -Scp option")
			}
			pos++
			all.Dep.ForceCP = args[pos]
		} else if args[pos] == "-Srepro" {
			all.Dep.Repro = true
		} else if args[pos] == "-Sforce" {
			all.Dep.Force = true
		} else if args[pos] == "-Spom" {
			all.Dep.Pom = true
		} else if args[pos] == "-Stree" {
			all.Dep.Tree = true
		} else if args[pos] == "-Sresolve-tags" {
			all.Dep.ResolveTags = true
		} else if args[pos] == "-Sverbose" {
			all.Dep.Verbose = true
		} else if args[pos] == "-Sdescribe" {
			all.Dep.Describe = true
		} else if args[pos] == "-Sthreads" {
			if all.Dep.Threads > 0 {
				return pos, errors.New("threads option " + args[pos] + " defined more than one time")
			}
			if pos+1 > len(args)-1 {
				return pos, errors.New("threads value (N) not defined for -Sthreads option")
			}
			pos++
			i, err := strconv.Atoi(args[pos])
			if err != nil {
				return pos, errors.New("threads value '" + args[pos] + "' is not a number")
			}
			all.Dep.Threads = i
		} else if args[pos] == "-Strace" {
			all.Dep.Trace = true
		} else if strings.HasPrefix(args[pos], "-S") {
			return pos, errors.New("invalid option:" + args[pos])
		} else {
			// move to the next options group
			break
		}

		pos++
	}

	return pos, nil
}

func setInitOpts(all *allOpts, args []string, pos int) (int, error) {
	for {
		if pos >= len(args) {
			break
		}

		if args[pos] == "-i" || args[pos] == "--init" {
			if len(all.Init.Init) > 0 {
				return pos, errors.New("init option " + args[pos] + " defined more than one time")
			}
			if pos+1 > len(args)-1 {
				return pos, errors.New("init path not defined for " + args[pos] + " option")
			}
			pos++
			all.Init.Init = args[pos]
		} else if args[pos] == "-e" || args[pos] == "--eval" {
			if len(all.Init.Eval) > 0 {
				return pos, errors.New("eval option " + args[pos] + " defined more than one time")
			}
			if pos+1 > len(args)-1 {
				return pos, errors.New("eval string not defined for " + args[pos] + " option")
			}
			pos++
			all.Init.Eval = args[pos]
		} else if args[pos] == "--report" {
			if len(all.Init.Report) > 0 {
				return pos, errors.New("report option " + args[pos] + " defined more than one time")
			}
			if pos+1 > len(args)-1 {
				return pos, errors.New("report target not defined for " + args[pos] + " option")
			}
			pos++
			if len(args[pos]) == 0 {
				return pos, errors.New("empty report target is not valid for " + args[pos-1] + " option")
			}
			all.Init.Report = args[pos]
		} else {
			// move to the next options group
			break
		}

		pos++
	}

	return pos, nil
}

func setMainOpts(all *allOpts, args []string, pos int) (int, error) {
	for {
		if pos >= len(args) {
			break
		}

		if args[pos] == "-m" || args[pos] == "--main" {
			if pos+1 > len(args)-1 {
				return pos, errors.New("main ns-name not defined for " + args[pos] + " option")
			}
			pos++
			all.Main.MainArgs = append(all.Main.MainArgs, "-m", args[pos])
		} else if args[pos] == "-r" || args[pos] == "--repl" {
			all.Main.Repl = true
		} else if args[pos] == "-h" || args[pos] == "-?" || args[pos] == "--help" {
			all.Main.Help = true
			all.Main.HelpArg = args[pos]
		} else {
			// get current option [should either be - or a path]
			// and move to the next options group
			all.Main.PathArg = args[pos]
		}
		pos++
		break
	}

	return pos, nil
}

func use(options *allOpts) error {
	// Execute resolve-tags command
	if options.Dep.ResolveTags {
		if fileExists("deps.edn") == false {
			return errors.New("deps.edn does not exist")
		}
		return start(resolveTagsCmd(toolsCp))
	}

	// Determine user config directory
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// If user config directory does not exist, create it
	// and copy there the example deps edn
	err = copyExampleDeps(configDir, tools4CljDir)
	if err != nil {
		return err
	}

	// Chain deps.edn in config paths. repro=skip config dir
	config.configProject = "deps.edn"
	configPaths := getConfigPaths(&config, configDir, tools4CljDir, options.Dep.Repro)

	// Determine whether to use user or project cache
	cacheDir := ""
	if fileExists("deps.edn") {
		cacheDir = ".cpcache"
	} else {
		// Determine user cache directory
		userCacheDir, err := getCljCacheDir(configDir)
		if err != nil {
			return err
		}
		cacheDir = userCacheDir
	}

	// Calculate a checksum based on current options and config paths
	ck := checksumOf(options, configPaths)

	// Build the file parameters:
	// libsFile, cpFile, jvmFile, mainFile
	buildCmdConfigs(&config, cacheDir, ck)

	if options.Dep.Verbose {
		fmt.Println("version      = " + version)
		fmt.Println("install_dir  = " + tools4CljDir)
		fmt.Println("config_dir   = " + configDir)
		fmt.Println("config_paths = " + join(configPaths, " "))
		fmt.Println("cache_dir    = " + cacheDir)
		fmt.Println("cp_file      = " + config.cpFile)
	}

	// Check for stale classpath
	stale, err := isStale(options, config, configPaths)
	if err != nil {
		return err
	}

	// Make tools args if needed
	buildToolsArgs(&config, stale, options)

	// If stale, run make-classpath to refresh cached classpath
	if stale && !options.Dep.Describe {
		if options.Dep.Verbose {
			fmt.Println("Refreshing classpath")
		}
		err := start(makeClassPathCmd(&config, toolsCp))
		if err != nil {
			return err
		}
	}

	// Get active classpath to use
	cp, err := activeClassPath(options, config)
	if err != nil {
		return err
	}

	// Finally...
	if options.Dep.Pom {
		err := start(generatePomCmd(&config, toolsCp))
		if err != nil {
			return err
		}
	} else if options.Dep.PrintClassPath {
		fmt.Println(cp)
	} else if options.Dep.Describe {
		pathVector := ""
		for _, configPath := range configPaths {
			pathVector += "\"" + configPath + "\" "
		}
		fmt.Println(argsDescription(pathVector, tools4CljDir, configDir, cacheDir, &config, options))
	} else if options.Dep.Tree {
		err := start(printTreeCmd(&config, toolsCp))
		if err != nil {
			return err
		}
	} else if options.Dep.Trace {
		fmt.Println("Writing trace.edn")
	} else {
		jvmCacheOpts := []string{}
		if fileExists(config.jvmFile) {
			b, err := ioutil.ReadFile(config.jvmFile)
			if err != nil {
				return err
			}
			jvmCacheOpts = strings.Split(string(b), " ")
		}
		mainCacheOpts := []string{}
		if fileExists(config.mainFile) {
			b, err := ioutil.ReadFile(config.mainFile)
			if err != nil {
				return err
			}
			mainCacheOpts = strings.Split(string(b), " ")
		}

		clojureArgs := []string{}
		clojureArgs = append(clojureArgs, getInitArgs(options)...)
		clojureArgs = append(clojureArgs, options.Main.MainArgs...)
		clojureArgs = append(clojureArgs, options.Main.PathArg)
		clojureArgs = append(clojureArgs, options.Args...)

		err := safeStart(clojureCmd(jvmCacheOpts, options.Dep.JvmOpts, config.libsFile,
			cp, mainCacheOpts, clojureArgs, options.Rlwrap))
		if err != nil {
			return err
		}
	}

	return nil
}

func checksumOf(options *allOpts, configPaths []string) string {
	prep := join([]string{
		options.Dep.ResolveAliases,
		options.Dep.ClassPathAliases,
		options.Dep.AllAliases,
		options.Dep.JvmAliases,
		options.Dep.MainAliases,
		options.Dep.DepsData}, "|")
	for _, v := range configPaths {
		if fileExists(v) {
			prep += "|" + v
		} else {
			prep += "|NIL"
		}
	}
	c := crc32.ChecksumIEEE([]byte(prep))
	return fmt.Sprintf("%d", c)
}

func isStale(options *allOpts, config t4cConfig, configPaths []string) (bool, error) {
	stale := false
	if options.Dep.Force || options.Dep.Trace || !fileExists(config.cpFile) {
		stale = true
	} else {
		for _, path := range configPaths {
			newer, err := checkIsNewerFile(path, config.cpFile)
			if err != nil {
				return false, err
			}
			if newer {
				stale = true
				break
			}
		}
	}
	return stale, nil
}

func buildToolsArgs(config *t4cConfig, stale bool, options *allOpts) {
	if stale || options.Dep.Pom {
		config.toolsArgs = []string{}
		if len(options.Dep.DepsData) > 0 {
			config.toolsArgs = append(config.toolsArgs, "--config-data", options.Dep.DepsData)
		}
		if len(options.Dep.ResolveAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-R"+options.Dep.ResolveAliases)
		}
		if len(options.Dep.ClassPathAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-C"+options.Dep.ClassPathAliases)
		}
		if len(options.Dep.JvmAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-J"+options.Dep.JvmAliases)
		}
		if len(options.Dep.MainAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-M"+options.Dep.MainAliases)
		}
		if len(options.Dep.AllAliases) > 0 {
			config.toolsArgs = append(config.toolsArgs, "-A"+options.Dep.AllAliases)
		}
		if len(options.Dep.ForceCP) > 0 {
			config.toolsArgs = append(config.toolsArgs, "--skip-cp")
		}
		if options.Dep.Threads > 0 {
			config.toolsArgs = append(config.toolsArgs, "--threads")
			config.toolsArgs = append(config.toolsArgs, strconv.Itoa(options.Dep.Threads))
		}
		if options.Dep.Trace {
			config.toolsArgs = append(config.toolsArgs, "--trace")
		}
	}
}

func activeClassPath(options *allOpts, config t4cConfig) (string, error) {
	var cp string
	if options.Dep.Describe {
		cp = ""
	} else if len(options.Dep.ForceCP) > 0 {
		cp = options.Dep.ForceCP
	} else {
		b, err := ioutil.ReadFile(config.cpFile)
		if err != nil {
			return "", err
		}
		cp = string(b)
	}
	return cp, nil
}

func argsDescription(pathVector string, toolsDir string, configDir string, cacheDir string, config *t4cConfig, options *allOpts) string {
	return `{:version "` + version + `"
 :config-files [` + pathVector + `]
 :config-user "` + config.configUser + `"
 :config-project "` + config.configProject + `"
 :install-dir "` + toolsDir + `"
 :config-dir "` + configDir + `"
 :cache-dir "` + cacheDir + `"
 :force ` + strconv.FormatBool(options.Dep.Force) + `
 :repro ` + strconv.FormatBool(options.Dep.Repro) + `
 :resolve-aliases "` + options.Dep.ResolveAliases + `"
 :classpath-aliases "` + options.Dep.ClassPathAliases + `"
 :jvm-aliases "` + options.Dep.JvmAliases + `"
 :main-aliases "` + options.Dep.MainAliases + `"
 :all-aliases "` + options.Dep.AllAliases + `"}`
}

func getInitArgs(options *allOpts) []string {
	initArgs := []string{}
	if options.Init.Init != "" {
		initArgs = append(initArgs, []string{`-i`, options.Init.Init}...)
	}
	if options.Init.Eval != "" {
		initArgs = append(initArgs, []string{`-e`, options.Init.Eval}...)
	}
	if options.Init.Report != "" {
		initArgs = append(initArgs, []string{`--report`, options.Init.Report}...)
	}
	return initArgs
}
