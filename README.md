# tools4clj


In the shadows of the official clojure tools:

```
const version = "1.10.1.492"
```


## What is this?

This is the **_go_ tools for clojure**. 

The _t4c_ installation produces two binaries that follow closely the functionality of **clojure/brew-install** command line scripts:
- the `clojure` binary launcher of the official clojure tools, and
- the `clj` binary, which uses `rlwrap` readline wrapper, for interactive repl use

Also, `clj --rebel` runs an extended version of clj that, instead of using `rlwrap`, it uses the terminal readline library **bhauman/rebel-readline**, giving a more polished look and feel. 

Note: on Windows *rebel* prompt, a long running function can not be stopped by Ctrl+C. Use the task manager... 

Although *rebel* extension looks great, it is a bit slow sometimes, so it may not always be the choice of heart. Thats why *rlwrap'ed clj* is the default.

Please, report any bugs at https://github.com/tasosx/tools4clj/issues, stating your platform, and steps to reproduce them, along with actual and expected results.

Plan is to keep this project up to date with _clojure/brew-install_ changes. 

Any requests/PRs for features are welcome, as long as they do not stray from the plan.

## Why this?

### Is there something different compared to official deps cli tools?

- An easy entrance to clojure world for go users. If you are a go user you are one command line away from clojured happiness.
- Resolve Windows quoting differences. No need to think quotes-translation-for-Windows. One clojure/clj tools command to run anywhere. So pick a published deps cli example on the internet and run it directly on Windows (99% of the deps-cli clojure examples are directed to POSIX users anyways).
- A pretty clojure repl. Use rebel when you want to view/display a prettier clojure dev UI.

### Is it only for Windows?

TL;DR. No, it is for any platform, but on Windows has some specific features.

For the Windows platform a decision was made, in order to mitigate the quote escaping mess (see: https://clojure.atlassian.net/browse/TDEPS-133, https://clojure.atlassian.net/browse/TDEPS-136), to prefer a unix oriented command line input, by accepting arguments enclosed in single or double quotes, like `'(print "test")'` or `"(print \"test\")"`. Windows double-quotes nesting is not supported. Unix single quotes nesting is not supported.

If you need to override this behaviour, and want to use the native Windows arguments parsing, use `--native-args` option. It is the default on all other platforms.


## Based upon...

Tried to keep the usage of this project's produced binaries inline with the official CLI tools, with the exception of the installation directory. This project uses `~/.tools4clj/[version]` folder for the installation of `deps.edn`, `example-deps.edn` and `clojure-tools-X.Y.Z.jar` files, and the `%GOPATH%/bin` folder for the binaries.

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

Looks and feel of `clj --rebel` are configurable through the `rebel-readline` options:

https://github.com/bhauman/rebel-readline/#config


## More

https://clojure.org/guides/getting_started

https://clojure.org/guides/deps_and_cli

https://clojure.org/reference/deps_and_cli#_usage

https://github.com/clojure/tools.cli

https://github.com/clojure/brew-install

https://github.com/hanslub42/rlwrap

https://github.com/bhauman/rebel-readline

https://golang.org/doc/code.html#GOPATH

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