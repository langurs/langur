# langur 0.18

[Langur](https://langurlang.org) is an open-source imperative/functional scripting language written in Go.

Langur source code is always UTF-8, with Linux line returns, no BOM, and no surrogate codes.

The recommended file extension is .langur.

At the start of a langur script file, you can use a shebang to specify the location of the interpreter (as you do with other types of script files).

Besides running script files, you can also use the REPL by building and running langur/repl/main.go. This is useful for development, as it allows you to print lexer tokens, parsed representations, compiled opcodes, and the VM result (all optional).

The revision history is in a separate file.

## requirements

I've compiled langur on Linux, using LiteIDE with Go 1.22. I've also compiled it on Windows using LiteIDE.

See installation instructions on the langurlang.org website.

## copyright and license

See LICENSE file.

