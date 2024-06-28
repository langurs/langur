Why a regexp package here?
==========================
As of langur 0.9.3, you don't have to copy files from the go/regexp package. 0.9.3 makes more changes and it is easier to just have the whole thing from the start.

This was copied from the Go 1.14.1 standard library regexp package. The purpose is to add functionality for langur.

1. replace with maximum count methods (not just ReplaceAll)
   The file regexp_replaceN.go was copied and modified from a portion the regexp/regexp.go file.
   It's a minor change that makes this work.
2. free-spacing mode with comments
   regexp/syntax/parse.go - see diff_parse.txt
   regexp/regexp.go - see diff_regexp.txt
   added free-spacing meta-characters to all_test.go
   left out extensive POSIX test files
   left out a couple of other tests b/c of lack of access to internal/testenv
   changed package internal import statements from "regexp/..." to "langur/regexp/..." so they would work together, instead of looking for the system regexp package

It would be nice if these changes could be incorporated into the standard library, as appropriate.

All the disclaimers for Go apply, as well as all the disclaimers for langur (see LICENSE file).

The following license information was copied from golang.org.

Copyright (c) 2009 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

