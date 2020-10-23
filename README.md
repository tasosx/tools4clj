# tools4clj


In the shadows of the official clojure tools:

```
const version = "1.10.1.727"
```


## What is this?

This is the **_go_ tools for clojure**. 

The _tools4clj_ build produces two binaries that follow closely the functionality of **clojure/brew-install** command line scripts:
- the `clojure` binary launcher of the official clojure tools, and
- the `clj` binary, which launches the official clojure tools within a `rlwrap` readline wrapper, intended for interactive repl use.

Plan is to keep this project up to date with _clojure/brew-install_ changes (focused on stable releases).

Any requests/PRs for features are welcome, as long as they do not stray from the plan.

Please, report any bugs at https://github.com/tasosx/tools4clj/issues, stating your platform, and steps to reproduce them, along with actual and expected results.

## Why this?

### Is there something different compared to official clojure/clj tools?

For a Go user:
- An easy entrance to clojure world. If you are a go user you are one command line away from clojured happiness.

For a Windows user:
- Resolve Windows (powershell/cmd/bash) quotes handling differences. One clojure/clj tools command to run anywhere. So pick a published deps clj/clojure command line example on the internet and run it on Windows, with no need to change/escape the quotes.

For any user:
- Same update procedure to all supported platforms.
- A pretty clojure repl. Use rebel when you want to view/display a prettier clojure dev UI.

### Is it only for Windows?

TL;DR. No, it is for any platform, but on Windows has some specific features.

For the Windows platform a decision was made, in order to mitigate the quote escaping mess (see: https://clojure.atlassian.net/browse/TDEPS-133, https://clojure.atlassian.net/browse/TDEPS-136), to prefer a unix oriented command line input, by accepting arguments enclosed in single or double quotes, like `'(print "test")'` or `"(print \"test\")"`. Windows double-quotes nesting or Unix single quotes nesting is not supported on Windows platform.

If you need to override this behaviour, and want to use the native Windows arguments parsing, use `--native-args` option. It is the default on all other platforms.


## Based upon...

Tried to keep the usage of this project's produced binaries inline with the official CLI tools, with the exception of the installation directory. This project uses `~/.tools4clj/[version]` folder for the installation of `deps.edn`, `example-deps.edn`, `exec.jar` and `clojure-tools-X.Y.Z.jar` files, and the `%GOPATH%/bin` folder for the binaries.

Check:
- https://clojure.org/reference/deps_and_cli for a Deps and CLI detailed reference
- https://clojure.org/guides/getting_started for the official installers (Mac, Linux, Windows and also building clojure from source)


## Usage

You need to have Java and, also, Go installed and ready to develop (...have your go path ready)

To install (or update) the `clojure` and `clj` binary launchers, on a shell/command line prompt, run:
```
go get -u github.com/tasosx/tools4clj/cmd/...
```

You are ready to _go_ clojure... 

For usage info try:
```
clojure --help
```

On the first launch of `clojure` or `clj`, after a tools4clj update or installation, it will download the latest referenced version of clojure tools. 

So an one-liner for a tools update is:
```
go get -u github.com/tasosx/tools4clj/cmd/... && clojure
```

Also, `clj --rebel` runs an extended version of clj that, instead of using `rlwrap`, it uses the terminal readline library **bhauman/rebel-readline**, giving a more polished look and feel. 

Note: on Windows *rebel* prompt, a long running function can not be stopped by Ctrl+C. Use the task manager... 

Although *rebel* extension looks great, it is a bit slow sometimes, so it may not always be the choice of heart. Thats why *rlwrap'ed clj* is the default.

Looks and feel of `clj --rebel` are configurable through the `rebel-readline` options:

https://github.com/bhauman/rebel-readline/#config


## More

### Clojure

https://clojure.org/guides/getting_started

### Clojure official CLI tools

https://clojure.org/guides/deps_and_cli

https://github.com/clojure/brew-install

### Readliners

https://github.com/hanslub42/rlwrap

https://github.com/bhauman/rebel-readline

### Go

https://golang.org/doc/code.html#GOPATH

### Windows "quotes" issue (fixed in tools4clj)

https://clojure.atlassian.net/browse/TDEPS-121

https://clojure.atlassian.net/browse/TDEPS-133

https://clojure.atlassian.net/browse/TDEPS-136


## Homepage

https://github.com/tasosx/tools4clj


## Copyright and License

Copyright (c) 2019 Tasos Mamaloukos.

All rights reserved. This program and the accompanying materials 
are made available under the terms of the Eclipse Public License v1.0
which accompanies this distribution.

The Eclipse Public License is available at
    https://www.eclipse.org/org/documents/epl-v10.html

SPDX-License-Identifier: EPL-1.0