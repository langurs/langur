Why a regexp package here?
==========================
This was copied from the Go 1.22.4 standard library regexp package. The purpose is to add functionality for langur.

It would be nice if replace with maximum count and free-spacing mode could be added to the standard library regexp package. The changes are very simple and shouldn't create any inefficiencies.

The changes made here (in langur/regexp) are as follows.
1. replace with maximum count method (not just ReplaceAll)
   new file regexp_replaceN.go with methods copied and modified from regexp/regexp.go

2. free-spacing mode with line comments; only a few small changes to the following files
   regexp/syntax/parse.go
   regexp/regexp.go

3. dropped all test files and also make_perl_groups.pl
   I'm sure they've been tested many times, and some of the tests don't want to run from where the files are (lack of access to internal/testenv).

4. changed imports from "regexp/syntax" to "langur/regexp/syntax" to look for the modified syntax package instead of the system package

All the disclaimers for Go apply, as well as all the disclaimers for langur (see LICENSE file). This is provided with ABSOLUTELY NO WARRANTY OF ANY KIND.

The following license information was copied from go.dev.

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

