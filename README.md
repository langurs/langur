# langur 0.15

[Langur](https://langurlang.org) is an open-source imperative/functional scripting language written in Go.

Langur source code is always UTF-8, with Linux line returns, no BOM, and no surrogate codes.

The recommended file extension is .langur.

At the start of a langur script file, you can use a shebang to specify the location of the interpreter (as you do with other types of script files).

Besides running script files, you can also use the REPL by building and running langur/repl/main.go. This is useful for development, as it allows you to print lexer tokens, parsed representations, compiled opcodes, and the VM result (all optional).

The revision history is in a separate file.

## requirements

I've compiled langur on Linux, using LiteIDE with Go 1.21. I've also compiled it on Windows using LiteIDE.

See installation instructions on the langurlang.org website.

Compiling this version of langur requires the following. As of 0.10.2, we're using Go modules, and your system might download and install golang.org/x/text automatically on first compilation.
+ [Unicode normalization library](https://godoc.org/golang.org/x/text/unicode/norm)

## copyright and license

See LICENSE file.

