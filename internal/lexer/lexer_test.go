// langur/lexer/lexer_test.go

package lexer

import (
	"langur/cpoint"
	"langur/str"
	"langur/token"
	"langur/trace"
	"testing"
)

func TestGeneralTokens(t *testing.T) {
	tests := []struct {
		input        string
		tok          *token.Token
		expectErrors bool
	}{
		{"", &token.Token{Type: token.EOF, Literal: ""}, false},

		{"α", &token.Token{Type: token.INVALID, Literal: "α"}, false},
		{"abc", &token.Token{Type: token.IDENT, Literal: "abc"}, false},
		{"_systemVar", &token.Token{Type: token.IDENT, Literal: "_systemVar"}, false},
		{"123", &token.Token{Type: token.INT, Literal: "123"}, false},
		{"1.23", &token.Token{Type: token.FLOAT, Literal: "1.23"}, false},
		{"16x123", &token.Token{Type: token.INT, Literal: "123", Code2: 16}, false},
		{"16x1.23", &token.Token{Type: token.FLOAT, Literal: "1.23", Code2: 16}, false},

		{"123i", &token.Token{Type: token.INT, Literal: "123", Code: token.CODE_IMAGINARY_NUMBER}, false},
		{"1.23i", &token.Token{Type: token.FLOAT, Literal: "1.23", Code: token.CODE_IMAGINARY_NUMBER}, false},
		{"1.23_i", &token.Token{Type: token.FLOAT, Literal: "1.23", Code: token.CODE_IMAGINARY_NUMBER}, false},

		{"-", &token.Token{Type: token.MINUS, Literal: "-"}, false},
		{"-123", &token.Token{Type: token.MINUS, Literal: "-"}, false},
		{"-1.23", &token.Token{Type: token.MINUS, Literal: "-"}, false},
		{"-abc", &token.Token{Type: token.MINUS, Literal: "-"}, false},

		{"16x", &token.Token{Type: token.INVALID, Literal: "16x", Code2: 16}, true},

		// good strings
		{`"abc"`, &token.Token{Type: token.STRING, Literal: "abc"}, false},
		{"qs'abc'", &token.Token{Type: token.STRING, Literal: "abc"}, false},
		{`qs"abc"`, &token.Token{Type: token.STRING, Literal: "abc"}, false},
		{"qs/abc/", &token.Token{Type: token.STRING, Literal: "abc"}, false},
		// {"qs(abc)", &token.Token{Type: token.STRING, Literal: "abc"}, false},
		// {"qs[abc]", &token.Token{Type: token.STRING, Literal: "abc"}, false},
		// {"qs<abc>", &token.Token{Type: token.STRING, Literal: "abc"}, false},
		{"QS'abc'", &token.Token{Type: token.STRING, Literal: "abc"}, false},
		{`QS"abc"`, &token.Token{Type: token.STRING, Literal: "abc"}, false},
		{"QS/abc/", &token.Token{Type: token.STRING, Literal: "abc"}, false},
		// {"QS(abc)", &token.Token{Type: token.STRING, Literal: "abc"}, false},
		// {"QS[abc]", &token.Token{Type: token.STRING, Literal: "abc"}, false},
		// {"QS<abc>", &token.Token{Type: token.STRING, Literal: "abc"}, false},

		// bad strings
		{`"abc`, nil, true},
		{"qs'abc", nil, true},
		{`qs"abc`, nil, true},
		{"qs/abc", nil, true},
		// {"qs(abc", nil, true},
		// {"qs[abc", nil, true},
		// {"qs<abc", nil, true},
		{"QS'abc", nil, true},
		{`QS"abc`, nil, true},
		{"QS/abc", nil, true},
		// {"QS(abc", nil, true},
		// {"QS[abc", nil, true},
		// {"QS<abc", nil, true},

		// good code point literals
		{"'a'", &token.Token{Type: token.INT, Literal: "97"}, false},
		{"'α'", &token.Token{Type: token.INT, Literal: "945"}, false},
		{"'\\''", &token.Token{Type: token.INT, Literal: "39"}, false},
		{"'\\n'", &token.Token{Type: token.INT, Literal: "10"}, false},
		{"'\\u000A'", &token.Token{Type: token.INT, Literal: "10"}, false},

		// bad code point literals
		{"'ab'", &token.Token{Type: token.INT, Literal: "97"}, true},
		{"'a", &token.Token{Type: token.INT, Literal: "97"}, true},
		{"'\n'", &token.Token{Type: token.INT, Literal: "10"}, true},
		{"''", &token.Token{Type: token.INVALID, Literal: ""}, true},

		// good re2 regex literals
		{"re'abc'", &token.Token{Type: token.REGEX_RE2, Literal: "(?-smiUx:abc)"}, false},
		{`re"abc"`, &token.Token{Type: token.REGEX_RE2, Literal: "(?-smiUx:abc)"}, false},
		{"re/abc/", &token.Token{Type: token.REGEX_RE2, Literal: "(?-smiUx:abc)"}, false},
		// {"re(abc)", &token.Token{Type: token.REGEX_RE2, Literal: "(?-smiUx:abc)"}, false},
		// {"re[abc]", &token.Token{Type: token.REGEX_RE2, Literal: "(?-smiUx:abc)"}, false},
		// {"re<abc>", &token.Token{Type: token.REGEX_RE2, Literal: "(?-smiUx:abc)"}, false},
		{"RE'abc'", &token.Token{Type: token.REGEX_RE2, Literal: "(?-smiUx:abc)"}, false},
		{`RE"abc"`, &token.Token{Type: token.REGEX_RE2, Literal: "(?-smiUx:abc)"}, false},
		{"RE/abc/", &token.Token{Type: token.REGEX_RE2, Literal: "(?-smiUx:abc)"}, false},
		// {"RE(abc)", &token.Token{Type: token.REGEX_RE2, Literal: "(?-smiUx:abc)"}, false},
		// {"RE[abc]", &token.Token{Type: token.REGEX_RE2, Literal: "(?-smiUx:abc)"}, false},
		// {"RE<abc>", &token.Token{Type: token.REGEX_RE2, Literal: "(?-smiUx:abc)"}, false},

		// bad re2 regex literals
		{"re'abc", nil, true},
		{`re"abc`, nil, true},
		{"re/abc", nil, true},
		// {"re(abc", nil, true},
		// {"re[abc", nil, true},
		// {"re<abc", nil, true},
		{"RE'abc", nil, true},
		{`RE"abc`, nil, true},
		{"RE/abc", nil, true},
		// {"RE(abc", nil, true},
		// {"RE[abc", nil, true},
		// {"RE<abc", nil, true},

		{"if", &token.Token{Type: token.IF, Literal: "if"}, false},

		{"(", &token.Token{Type: token.LPAREN, Literal: "("}, false},
		{")", &token.Token{Type: token.RPAREN, Literal: ")"}, false},

		{"{", &token.Token{Type: token.LBRACE, Literal: "{"}, false},
		{"}", &token.Token{Type: token.RBRACE, Literal: "}"}, false},

		{"[", &token.Token{Type: token.LBRACKET, Literal: "["}, false},
		{"]", &token.Token{Type: token.RBRACKET, Literal: "]"}, false},

		{"<", &token.Token{Type: token.LESS_THAN, Literal: "<"}, false},
		{">", &token.Token{Type: token.GREATER_THAN, Literal: ">"}, false},
		{"<=", &token.Token{Type: token.LT_OR_EQUAL, Literal: "<="}, false},
		{">=", &token.Token{Type: token.GT_OR_EQUAL, Literal: ">="}, false},
		{"<?", &token.Token{Type: token.LESS_THAN, Literal: "<?", Code: token.CODE_DB_OPERATOR}, false},
		{">?", &token.Token{Type: token.GREATER_THAN, Literal: ">?", Code: token.CODE_DB_OPERATOR}, false},
		{"<=?", &token.Token{Type: token.LT_OR_EQUAL, Literal: "<=?", Code: token.CODE_DB_OPERATOR}, false},
		{">=?", &token.Token{Type: token.GT_OR_EQUAL, Literal: ">=?", Code: token.CODE_DB_OPERATOR}, false},

		{"*", &token.Token{Type: token.ASTERISK, Literal: "*"}, false},
		{"/", &token.Token{Type: token.SLASH, Literal: "/"}, false},
		{`\`, &token.Token{Type: token.BACKSLASH, Literal: `\`}, false},
		{"rem", &token.Token{Type: token.REMAINDER, Literal: "rem"}, false},
		{"mod", &token.Token{Type: token.MODULUS, Literal: "mod"}, false},
		{"+", &token.Token{Type: token.PLUS, Literal: "+"}, false},
		{"-", &token.Token{Type: token.MINUS, Literal: "-"}, false},
		{":", &token.Token{Type: token.COLON, Literal: ":"}, false},
		{";", &token.Token{Type: token.SEMICOLON, Literal: ";"}, false},
		{",", &token.Token{Type: token.COMMA, Literal: ","}, false},
		{"..", &token.Token{Type: token.RANGE, Literal: ".."}, false},
		{"_", &token.Token{Type: token.NONE, Literal: "_"}, false},

		{".", &token.Token{Type: token.DOT, Literal: "."}, false},

		{"=", &token.Token{Type: token.ASSIGN, Literal: "="}, false},
		{"var", &token.Token{Type: token.VAR, Literal: "var"}, false},
		{"val", &token.Token{Type: token.VAL, Literal: "val"}, false},

		{"==", &token.Token{Type: token.EQUAL, Literal: "=="}, false},
		{"!=", &token.Token{Type: token.NOT_EQUAL, Literal: "!="}, false},
		{"==?", &token.Token{Type: token.EQUAL, Literal: "==?", Code: token.CODE_DB_OPERATOR}, false},
		{"!=?", &token.Token{Type: token.NOT_EQUAL, Literal: "!=?", Code: token.CODE_DB_OPERATOR}, false},

		{"->", &token.Token{Type: token.FORWARD, Literal: "->"}, false},

		{"and", &token.Token{Type: token.AND, Literal: "and"}, false},
		{"or", &token.Token{Type: token.OR, Literal: "or"}, false},
		{"nand", &token.Token{Type: token.NAND, Literal: "nand"}, false},
		{"nor", &token.Token{Type: token.NOR, Literal: "nor"}, false},
		{"xor", &token.Token{Type: token.XOR, Literal: "xor"}, false},
		{"nxor", &token.Token{Type: token.NXOR, Literal: "nxor"}, false},
		{"not", &token.Token{Type: token.NOT, Literal: "not"}, false},
		{"and?", &token.Token{Type: token.AND, Literal: "and?", Code: token.CODE_DB_OPERATOR}, false},
		{"or?", &token.Token{Type: token.OR, Literal: "or?", Code: token.CODE_DB_OPERATOR}, false},
		{"nand?", &token.Token{Type: token.NAND, Literal: "nand?", Code: token.CODE_DB_OPERATOR}, false},
		{"nor?", &token.Token{Type: token.NOR, Literal: "nor?", Code: token.CODE_DB_OPERATOR}, false},
		{"xor?", &token.Token{Type: token.XOR, Literal: "xor?", Code: token.CODE_DB_OPERATOR}, false},
		{"nxor?", &token.Token{Type: token.NXOR, Literal: "nxor?", Code: token.CODE_DB_OPERATOR}, false},
		{"not?", &token.Token{Type: token.NOT, Literal: "not?", Code: token.CODE_DB_OPERATOR}, false},

		{"# line comment", &token.Token{Type: token.EOF, Literal: ""}, false},
		{"/* inline comment */", &token.Token{Type: token.EOF, Literal: ""}, false},
		{"/* failed (incomplete) inline comment", &token.Token{Type: token.EOF, Literal: ""}, true},

		{"zls", &token.Token{Type: token.STRING, Literal: ""}, false},
	}

	// fmt.Println("System Newline : " + constants["N"])

	for _, tt := range tests {
		l, err := New(tt.input, "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		tok, err := l.NextToken()
		if err != nil && !tt.expectErrors {
			t.Fatal(err.Error())
		}

		if tt.tok != nil {
			if tok.Type != tt.tok.Type {
				t.Errorf("(%q) token type expected=%s, received=%s", tt.input, token.TypeDescription(tt.tok.Type), token.TypeDescription(tok.Type))
			}

			if tok.Literal != tt.tok.Literal {
				t.Errorf("(%q) token literal expected=%q, received=%q", tt.input, tt.tok.Literal, tok.Literal)
			}

			if tok.Code != tt.tok.Code {
				t.Errorf("(%q) token Code expected=%d, received=%d", tt.input, tt.tok.Code, tok.Code)
			}

			if tok.Code2 != tt.tok.Code2 {
				t.Errorf("(%q) token Code2 expected=%d, received=%d", tt.input, tt.tok.Code2, tok.Code2)
			}
		}

		if !tt.expectErrors && tok.Errs != nil {
			for _, e := range tok.Errs {
				t.Errorf("Lex Error (Count: %d): %s", e.Count, e.Err.Error())
			}
		} else if tt.expectErrors && tok.Errs == nil {
			t.Errorf("(%q) token expected to generate error, but no error generated", tt.input)
		}
	}
}

func TestStringLiterals(t *testing.T) {
	tests := []struct {
		input        string
		expected     string
		expectErrors bool
	}{
		{`""`, "", false},
		{`"abc"`, "abc", false},
		{`qs"abc"`, "abc", false},
		{`qs'abc'`, "abc", false},
		{`qs/abc/`, "abc", false},
		// {`qs(abc)`, "abc", false},
		// {`qs[abc]`, "abc", false},
		// {`qs<abc>`, "abc", false},

		// blockquotes
		{"qs:block END\nabc\nEND\n", "abc", false},
		{"qs:block GREAT_MARKER_EH\nabc\n123\nGREAT_MARKER_EH\n", "abc\n123", false},
		{"qs:block END\nabc\n  END\n", "abc", false},

		// {`"\N"`, str.SysNewLine, false},
		{`qs'\tyo שלם\t'`, "\tyo שלם\t", false},
		{`qs"\tabc"`, "\tabc", false},
		{"qs'\\t\\n'", "\t\n", false},
		{"QS'\\t\\n'", "\\t\\n", false},

		// good escape sequences
		// {`"\'"`, "'", false},
		{`"\\"`, "\\", false},
		{`"\r"`, "\r", false},
		{`"\n"`, "\n", false},
		// {`"\L"`, "\u2028", false},
		// {`"\P"`, "\u2029", false},
		{`"\t"`, "\t", false},
		{`"\0"`, "\000", false},
		{`"\e"`, "\x1B", false},

		{`"\uFEFF"`, "\ufeff", false},
		{`"\u0085"`, "\u0085", false},
		{`"\x0F"`, "\x0f", false},
		{`"\o100"`, "\x40", false},
		{`"abc\uFFFF123"`, "abc\uffff123", false},
		{`"abc\x33123"`, "abc\x33123", false},
		{`"abc\o123123"`, "abc\123123", false},

		{`"\U0010FFFF"`, "\U0010ffff", false},

		// bad escape sequences
		{`"\+"`, string(cpoint.REPLACEMENT), true},
		{`"\xFF"`, string(cpoint.REPLACEMENT), true},
		{`"\o377"`, string(cpoint.REPLACEMENT), true},
		{`"\xF"`, string(cpoint.REPLACEMENT), true},
		{`"\uF"`, string(cpoint.REPLACEMENT), true},
		{`"\o1"`, string(cpoint.REPLACEMENT), true},
		{`"\oA"`, string(cpoint.REPLACEMENT), true},
		{`"abc\+123"`, "abc" + string(cpoint.REPLACEMENT) + "123", true},
		{`"abc\uGfff123"`, "abc" + string(cpoint.REPLACEMENT) + "123", true},
		{`"abc\xGF123"`, "abc" + string(cpoint.REPLACEMENT) + "123", true},
		{`"abc\o888123"`, "abc" + string(cpoint.REPLACEMENT) + "123", true},
	}

	for _, tt := range tests {
		l, err := New(tt.input, "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		tok, err := l.NextToken()
		if err != nil && !tt.expectErrors {
			t.Fatal(err.Error())
		}

		if tok.Type != token.STRING {
			t.Fatalf("(%q) expected string token, received=%s", tt.input, token.TypeDescription(tok.Type))
		}

		if tok.Literal != tt.expected {
			t.Errorf("string literal result expected=%q, received=%q", tt.expected, tok.Literal)
		}

		if !tt.expectErrors && tok.Errs != nil {
			for _, e := range tok.Errs {
				t.Errorf("Lex Error (Count: %d): %s", e.Count, e.Err.Error())
			}
		} else if tt.expectErrors && tok.Errs == nil {
			t.Errorf("(%q) token expected to generate error, but no error generated", tt.input)
		}

	}
}

func TestDateTimeLiterals(t *testing.T) {
	tests := []struct {
		input        string
		expected     string
		expectErrors bool
	}{
		// lexer just reads in the text; doesn't decode / verify formatting
		{`dt""`, "", false},
		{`dt//`, "", false},
		{`dt/2018-01-01T12:01:00/`, "2018-01-01T12:01:00", false},
		{`dt/2018-01-01 12:01:00/`, "2018-01-01 12:01:00", false},

		// bad literals
		{`dt`, "", true},
		{`dt //`, "", true},
		{`dt/2018-01-01 12:01:00`, "2018-01-01 12:01:00", true},
	}

	for _, tt := range tests {
		l, err := New(tt.input, "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		tok, err := l.NextToken()
		if err != nil && !tt.expectErrors {
			t.Fatal(err.Error())
		}

		if !tt.expectErrors && tok.Type != token.DATETIME {
			t.Fatalf("(%q) expected date-time token, received=%s", tt.input, token.TypeDescription(tok.Type))
		}

		if tok.Literal != tt.expected {
			t.Errorf("date-time literal result expected=%q, received=%q", tt.expected, tok.Literal)
		}

		if !tt.expectErrors && tok.Errs != nil {
			for _, e := range tok.Errs {
				t.Errorf("Lex Error (Count: %d): %s", e.Count, e.Err.Error())
			}
		} else if tt.expectErrors && tok.Errs == nil {
			t.Errorf("(%q) token expected to generate error, but no error generated", tt.input)
		}

	}
}

func TestStringBalance(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// LTR/RTL embedding/isolate marks balanced?
		{"", true},
		{"abcd", true},
		{"\u2066abcd\u2069", true},
		{"\u2066abcd", false},
		{"abcd\u2069", false},

		{"\u2067abcd\u2069", true},
		{"\u2067abcd", false},
		{"abcd\u2069", false},

		{"\u2068abcd\u2069", true},
		{"\u2068abcd", false},
		{"abcd\u2069", false},
		{"abcd\u202C", false},

		{"\u2066a\u202Abc\u202Cd\u2069", true},
		{"\u2066a\u202Abc\u2069d\u202C", false},

		{"\u202Aa\u2066bc\u2069d\u202C", true},
		{"\u202Aa\u2066bc\u202Cd\u2069", false},
	}

	for _, tt := range tests {
		lex, err := New("qs:any\""+tt.input+"\"", "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		tok, err := lex.NextToken()
		if err != nil {
			t.Fatal(err.Error())
		}

		if (tok.Errs == nil) != tt.expected {
			t.Errorf("String balanced expected to be %t, received=%t", tt.expected, tok.Errs == nil)
		}
	}
}

func TestTokenSequencing(t *testing.T) {
	// Do tokens work in sequence?

	tests := []struct {
		input    string
		expected []*token.Token
	}{
		{`"asdf"
			-123`,
			[]*token.Token{
				&token.Token{Type: token.STRING, Literal: "asdf"},
				&token.Token{Type: token.SEMICOLON, Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL},
				&token.Token{Type: token.MINUS, Literal: "-"},
				&token.Token{Type: token.INT, Literal: "123"},
				&token.Token{Literal: "", Type: token.EOF},
			}},

		{`"asdf"
			catch`,
			[]*token.Token{
				&token.Token{Type: token.STRING, Literal: "asdf"},
				&token.Token{Type: token.SEMICOLON, Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL},
				&token.Token{Type: token.CATCH, Literal: "catch"},
				&token.Token{Literal: "", Type: token.EOF},
			}},

		{`fn() { ""
				45
				catch { case 7 : 890 }
				"asdf"
				catch { case 89: 0 }
				}`,
			[]*token.Token{
				&token.Token{Type: token.FUNCTION, Literal: "fn"},
				&token.Token{Type: token.LPAREN, Literal: "("},
				&token.Token{Type: token.RPAREN, Literal: ")"},
				&token.Token{Type: token.LBRACE, Literal: "{"},
				&token.Token{Type: token.STRING, Literal: ""},
				&token.Token{Type: token.SEMICOLON, Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL},
				&token.Token{Type: token.INT, Literal: "45"},
				&token.Token{Type: token.SEMICOLON, Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL},
				&token.Token{Type: token.CATCH, Literal: "catch"},
				&token.Token{Type: token.LBRACE, Literal: "{"},
				&token.Token{Type: token.CASE, Literal: "case"},
				&token.Token{Type: token.INT, Literal: "7"},
				&token.Token{Type: token.COLON, Literal: ":"},
				&token.Token{Type: token.INT, Literal: "890"},
				&token.Token{Type: token.RBRACE, Literal: "}"},
				&token.Token{Type: token.SEMICOLON, Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL},
				&token.Token{Type: token.STRING, Literal: "asdf"},
				&token.Token{Type: token.SEMICOLON, Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL},
				&token.Token{Type: token.CATCH, Literal: "catch"},
				&token.Token{Type: token.LBRACE, Literal: "{"},
				&token.Token{Type: token.CASE, Literal: "case"},
				&token.Token{Type: token.INT, Literal: "89"},
				&token.Token{Type: token.COLON, Literal: ":"},
				&token.Token{Type: token.INT, Literal: "0"},
				&token.Token{Type: token.RBRACE, Literal: "}"},
				&token.Token{Type: token.SEMICOLON, Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL},
				&token.Token{Type: token.RBRACE, Literal: "}"},
			},
		},

		{"fn{*}", []*token.Token{
			&token.Token{Type: token.FUNCTION, Literal: "fn"},
			&token.Token{Type: token.LBRACE, Literal: "{"},
			&token.Token{Type: token.ASTERISK, Literal: "*"},
			&token.Token{Type: token.RBRACE, Literal: "}"},
		}},

		{"fn{* 2}", []*token.Token{
			&token.Token{Type: token.FUNCTION, Literal: "fn"},
			&token.Token{Type: token.LBRACE, Literal: "{"},
			&token.Token{Type: token.ASTERISK, Literal: "*"},
			&token.Token{Type: token.INT, Literal: "2"},
			&token.Token{Type: token.RBRACE, Literal: "}"},
		}},

		{"fn{+2}", []*token.Token{
			&token.Token{Type: token.FUNCTION, Literal: "fn"},
			&token.Token{Type: token.LBRACE, Literal: "{"},
			&token.Token{Type: token.PLUS, Literal: "+"},
			&token.Token{Type: token.INT, Literal: "2"},
			&token.Token{Type: token.RBRACE, Literal: "}"},
		}},

		{"fn{-2}", []*token.Token{
			&token.Token{Type: token.FUNCTION, Literal: "fn"},
			&token.Token{Type: token.LBRACE, Literal: "{"},
			&token.Token{Type: token.MINUS, Literal: "-"},
			&token.Token{Type: token.INT, Literal: "2"},
			&token.Token{Type: token.RBRACE, Literal: "}"},
		}},

		{"-123", []*token.Token{
			&token.Token{Type: token.MINUS, Literal: "-"},
			&token.Token{Type: token.INT, Literal: "123"},
		}},

		{"- 123", []*token.Token{
			&token.Token{Type: token.MINUS, Literal: "-"},
			&token.Token{Type: token.INT, Literal: "123"},
		}},

		{`i -123`,
			[]*token.Token{
				&token.Token{Literal: "i", Type: token.IDENT},
				&token.Token{Type: token.MINUS, Literal: "-"},
				&token.Token{Type: token.INT, Literal: "123"},
			}},

		{`i-123`,
			[]*token.Token{
				&token.Token{Literal: "i", Type: token.IDENT},
				&token.Token{Literal: "-", Type: token.MINUS},
				&token.Token{Type: token.INT, Literal: "123"},
			}},

		{`i - 123`,
			[]*token.Token{
				&token.Token{Literal: "i", Type: token.IDENT},
				&token.Token{Literal: "-", Type: token.MINUS},
				&token.Token{Type: token.INT, Literal: "123"},
			}},

		{`12-123`,
			[]*token.Token{
				&token.Token{Type: token.INT, Literal: "12"},
				&token.Token{Literal: "-", Type: token.MINUS},
				&token.Token{Type: token.INT, Literal: "123"},
			}},

		{"-ABC", []*token.Token{
			&token.Token{Type: token.MINUS, Literal: "-"},
			&token.Token{Type: token.IDENT, Literal: "ABC"},
		}},

		{`i + -123`,
			[]*token.Token{
				&token.Token{Literal: "i", Type: token.IDENT},
				&token.Token{Literal: "+", Type: token.PLUS},
				&token.Token{Type: token.MINUS, Literal: "-"},
				&token.Token{Type: token.INT, Literal: "123"},
			}},

		{"abc[123]",
			[]*token.Token{
				&token.Token{Literal: "abc", Type: token.IDENT},
				&token.Token{Literal: "[", Type: token.LBRACKET},
				&token.Token{Literal: "123", Type: token.INT},
				&token.Token{Literal: "]", Type: token.RBRACKET},
			},
		},

		{"abc'A123",
			[]*token.Token{
				&token.Token{Literal: "abc", Type: token.IDENT},
				&token.Token{Literal: "([)", Type: token.LBRACKET},
				&token.Token{Literal: "A123", Type: token.STRING},
				&token.Token{Literal: "(])", Type: token.RBRACKET},
			},
		},

		{"abc'123",
			[]*token.Token{
				&token.Token{Literal: "abc", Type: token.IDENT},
				&token.Token{Literal: "([)", Type: token.LBRACKET},
				&token.Token{Literal: "123", Type: token.STRING},
				&token.Token{Literal: "(])", Type: token.RBRACKET},
			},
		},

		{`"s1""s2"`,
			[]*token.Token{
				&token.Token{Literal: "s1", Type: token.STRING},
				&token.Token{Literal: "s2", Type: token.STRING},
			},
		},

		{`"s1" "s2"`,
			[]*token.Token{
				&token.Token{Literal: "s1", Type: token.STRING},
				&token.Token{Literal: "s2", Type: token.STRING},
			},
		},

		{"not abc",
			[]*token.Token{
				&token.Token{Type: token.NOT, Literal: "not"},
				&token.Token{Literal: "abc", Type: token.IDENT},
			},
		},

		{"testy# line comment\nabc",
			[]*token.Token{
				&token.Token{Literal: "testy", Type: token.IDENT},
				&token.Token{Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL, Type: token.SEMICOLON},
				&token.Token{Literal: "abc", Type: token.IDENT},
			},
		},

		{"a#\n234",
			[]*token.Token{
				&token.Token{Literal: "a", Type: token.IDENT},
				&token.Token{Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL, Type: token.SEMICOLON},
				&token.Token{Literal: "234", Type: token.INT},
			},
		},

		{"==/* inline \n comment */ abc",
			[]*token.Token{
				&token.Token{Literal: "==", Type: token.EQUAL},
				&token.Token{Literal: "abc", Type: token.IDENT},
			},
		},

		{"if true {\n 1 } else { 2 }\n 10 == 11",
			[]*token.Token{
				&token.Token{Literal: "if", Type: token.IF},
				&token.Token{Literal: "true", Type: token.TRUE},
				&token.Token{Literal: "{", Type: token.LBRACE},
				&token.Token{Literal: "1", Type: token.INT},
				&token.Token{Literal: "}", Type: token.RBRACE},
				&token.Token{Literal: "else", Type: token.ELSE},
				&token.Token{Literal: "{", Type: token.LBRACE},
				&token.Token{Literal: "2", Type: token.INT},
				&token.Token{Literal: "}", Type: token.RBRACE},
				&token.Token{Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL, Type: token.SEMICOLON},
				&token.Token{Literal: "10", Type: token.INT},
				&token.Token{Literal: "==", Type: token.EQUAL},
				&token.Token{Literal: "11", Type: token.INT},
			},
		},

		{"fw/don't you know/",
			[]*token.Token{
				&token.Token{Literal: "([)", Type: token.LBRACKET},
				&token.Token{Literal: "don't", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "you", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "know", Type: token.STRING},
				&token.Token{Literal: "(])", Type: token.RBRACKET},
			},
		},
		{`fw/don't\x20you know/`,
			[]*token.Token{
				&token.Token{Literal: "([)", Type: token.LBRACKET},
				&token.Token{Literal: "don't you", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "know", Type: token.STRING},
				&token.Token{Literal: "(])", Type: token.RBRACKET},
			},
		},
		{`FW/don't\x20you know/`,
			[]*token.Token{
				&token.Token{Literal: "([)", Type: token.LBRACKET},
				&token.Token{Literal: `don't\x20you`, Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "know", Type: token.STRING},
				&token.Token{Literal: "(])", Type: token.RBRACKET},
			},
		},

		{"fw/ don't you know/", // leading space
			[]*token.Token{
				&token.Token{Literal: "([)", Type: token.LBRACKET},
				&token.Token{Literal: "don't", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "you", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "know", Type: token.STRING},
				&token.Token{Literal: "(])", Type: token.RBRACKET},
			},
		},
		{"fw/don't you know /", // trailing space
			[]*token.Token{
				&token.Token{Literal: "([)", Type: token.LBRACKET},
				&token.Token{Literal: "don't", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "you", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "know", Type: token.STRING},
				&token.Token{Literal: "(])", Type: token.RBRACKET},
			},
		},
		{"fw/ don't you know /", // leading and trailing space
			[]*token.Token{
				&token.Token{Literal: "([)", Type: token.LBRACKET},
				&token.Token{Literal: "don't", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "you", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "know", Type: token.STRING},
				&token.Token{Literal: "(])", Type: token.RBRACKET},
			},
		},
		{"fw/ don't \tyou   know /", // and more spaces
			[]*token.Token{
				&token.Token{Literal: "([)", Type: token.LBRACKET},
				&token.Token{Literal: "don't", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "you", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "know", Type: token.STRING},
				&token.Token{Literal: "(])", Type: token.RBRACKET},
			},
		},
		{`fw:block WORDS
		don't  you
		know
		WORDS`,
			[]*token.Token{
				&token.Token{Literal: "([)", Type: token.LBRACKET},
				&token.Token{Literal: "don't", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "you", Type: token.STRING},
				&token.Token{Literal: "(,)", Type: token.COMMA},
				&token.Token{Literal: "know", Type: token.STRING},
				&token.Token{Literal: "(])", Type: token.RBRACKET},
				&token.Token{Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL, Type: token.SEMICOLON},
			},
		},
	}

	for _, tt := range tests {
		l, err := New(tt.input, "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}

		for _, expectToken := range tt.expected {
			tok, err := l.NextToken()
			if err != nil {
				t.Fatal(err.Error())
			}

			if expectToken.Type != tok.Type {
				t.Errorf("(%q) token type expected=%s, received=%s (%q)", tt.input, token.TypeDescription(expectToken.Type), token.TypeDescription(tok.Type), tok.Literal)
			}

			if expectToken.Literal != tok.Literal {
				t.Errorf("(%q) token literal expected=%q, received=%q", tt.input, expectToken.Literal, tok.Literal)
			}

			if expectToken.Code != tok.Code {
				t.Errorf("(%q) token code expected=%d, received=%d", tt.input, expectToken.Code, tok.Code)
			}

			if tok.Errs != nil {
				for _, e := range tok.Errs {
					t.Errorf("(%q; %q) Lex Error (%s): %s", tt.input, tok.Literal, str.IntToStr(e.Count, 10), e.Err.Error())
				}
			}
		}
		tok, err := l.NextToken()
		if err != nil {
			t.Fatal(err.Error())
		}
		if tok.Type != token.EOF {
			t.Errorf("(%q) More tokens beyond expected tokens\nNext token: %s", tt.input, tok.String())
		}
	}
}

// check line number reporting as well
func TestTokenSequencingLineReporting(t *testing.T) {
	tests := []struct {
		input    string
		expected []*token.Token
	}{
		{`"s1""s2"`,
			[]*token.Token{
				&token.Token{Literal: "s1", Type: token.STRING,
					Where: trace.NewWhere(1, 1),
				},
				&token.Token{Literal: "s2", Type: token.STRING,
					Where: trace.NewWhere(1, 5),
				},
			},
		},

		{` s1`,
			[]*token.Token{
				&token.Token{Literal: "s1", Type: token.IDENT,
					Where: trace.NewWhere(1, 2),
				},
			},
		},

		{`s1 s2`,
			[]*token.Token{
				&token.Token{Literal: "s1", Type: token.IDENT,
					Where: trace.NewWhere(1, 1),
				},
				&token.Token{Literal: "s2", Type: token.IDENT,
					Where: trace.NewWhere(1, 4),
				},
			},
		},

		{`s1
s2`,
			[]*token.Token{
				&token.Token{Literal: "s1", Type: token.IDENT,
					Where: trace.NewWhere(1, 1),
				},
				&token.Token{Type: token.SEMICOLON, Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL,
					Where: trace.NewWhere(1, 3),
				},
				&token.Token{Literal: "s2", Type: token.IDENT,
					Where:           trace.NewWhere(2, 1),
					NewLinePrecedes: true,
				},
			},
		},
	}

	for _, tt := range tests {
		l, err := New(tt.input, "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}

		for _, expectToken := range tt.expected {
			tok, err := l.NextToken()
			if err != nil {
				t.Fatal(err.Error())
			}

			if expectToken.Type != tok.Type {
				t.Errorf("(%q) token type expected=%s, received=%s (%q)", tt.input, token.TypeDescription(expectToken.Type), token.TypeDescription(tok.Type), tok.Literal)
			}

			if expectToken.Literal != tok.Literal {
				t.Errorf("(%q) token literal expected=%q, received=%q", tt.input, expectToken.Literal, tok.Literal)
			}

			if expectToken.Code != tok.Code {
				t.Errorf("(%q) token code expected=%d, received=%d", tt.input, expectToken.Code, tok.Code)
			}

			if expectToken.Where.Line != tok.Where.Line || expectToken.Where.LinePosition != tok.Where.LinePosition {
				t.Errorf("(%q) token line/position expected=%d/%d, received=%d/%d",
					tt.input, expectToken.Where.Line, expectToken.Where.LinePosition,
					tok.Where.Line, tok.Where.LinePosition)
			}

			if expectToken.NewLinePrecedes != tok.NewLinePrecedes {
				t.Errorf("(%q) token NewLinePrecedes expected=%t, received=%t",
					tt.input, expectToken.NewLinePrecedes, tok.NewLinePrecedes)
			}

			if tok.Errs != nil {
				for _, e := range tok.Errs {
					t.Errorf("(%q; %q) Lex Error (%s): %s", tt.input, tok.Literal, str.IntToStr(e.Count, 10), e.Err.Error())
				}
			}
		}
		tok, err := l.NextToken()
		if err != nil {
			t.Fatal(err.Error())
		}
		if tok.Type != token.EOF {
			t.Errorf("(%q) More tokens beyond expected tokens\nNext token: %s", tt.input, tok.String())
		}
	}
}
