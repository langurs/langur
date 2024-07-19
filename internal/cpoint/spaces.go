// langur/cpoint/spaces.go

package cpoint

// spacing code points allowed in langur source code

import (
	"unicode"
)

func IsTrimmable(cp rune) bool {
	return unicode.IsSpace(cp)
}

func IsTrimmableLeadingSpace(cp rune) bool {
	return unicode.IsSpace(cp) && !IsVerticalSpace(cp)
}

func IsTokenSpace(cp rune) bool {
	return IsTokenHorizontalSpace(cp) || IsTokenVerticalSpace(cp)
}

func IsTokenHorizontalSpace(cp rune) bool {
	// ASCII space or tab
	return cp == 0x20 || cp == 0x09
}

func IsTokenVerticalSpace(cp rune) bool {
	// ASCII line feed
	return cp == 0x0A
}

func IsVerticalSpace(cp rune) bool {
	return cp >= 0x0A && cp <= 0x0D ||
		cp == 0x85 ||
		cp == 0x2028 || cp == 0x2029
}

func IsPrivateUse(cp rune) bool {
	return cp >= 0xE000 && cp <= 0xF8FF ||
		cp >= 0xF0000 && cp <= 0xFFFFD ||
		cp >= 0x100000 && cp <= 0x10FFFD
}

func IsAllowedInComments(cp rune) bool {
	return IsAllowed(cp, true, false, true)
}

func IsAllowedInQuotedWordLiterals(cp rune) bool {
	if IsAllowedSpace(cp) && !IsTokenSpace(cp) {
		return false
	}
	return IsAllowed(cp, false, false, false)
}

func IsAllowedInStringLiterals(cp rune) bool {
	return IsAllowed(cp, false, false, false)
}

func IsAllowedSpace(cp rune) bool {
	return cp == 0x20 || cp == '\n' || cp == '\t'
}

func IsAllowed(cp rune, pua, allVertSpc, invisibleSpc bool) bool {
	if IsVerticalSpace(cp) {
		if allVertSpc {
			return true
		}
		return IsTokenVerticalSpace(cp)
		// does not account for all instances, such as strings that don't allow ANY vertical spacing (handled in the lexer)
	}
	if unicode.IsGraphic(cp) || IsAllowedSpace(cp) {
		return true
	}
	if pua && IsPrivateUse(cp) {
		return true
	}
	if invisibleSpc && IsAllowableInvisibleSpace(cp) {
		return true
	}
	return false
}

func IsAllowableInvisibleSpace(cp rune) bool {
	// space code points not included in the Unicode Graphic or Space categories
	// optionally allowed on string and regex literals
	switch cp {
	case
		// excluding FEFF, since its use as a space is deprecated
		0x180E, // MONGOLIAN VOWEL SEPARATOR

		0x200B, // ZERO WIDTH SPACE
		0x200C, // ZERO WIDTH NON-JOINER
		0x200D, // ZERO WIDTH JOINER
		0x2060, // WORD JOINER

		// DIRECTIONAL MARKERS ...
		0x200E, // LEFT-TO-RIGHT MARK
		0x200F, // RIGHT-TO-LEFT MARK
		0x061C, // ARABIC LETTER MARK

		// DIRECTIONAL MARKERS REQUIRING BALANCED CODE POINTS ...
		0x2066, // LEFT-TO-RIGHT ISOLATE
		0x2067, // RIGHT-TO-LEFT ISOLATE
		0x2068, // FIRST-STRONG ISOLATE
		0x2069, // POP DIRECTIONAL ISOLATE

		0x202A, // LEFT-TO-RIGHT EMBEDDING
		0x202B, // RIGHT-TO-LEFT EMBEDDING
		0x202C, // POP DIRECTIONAL FORMATTING
		0x202D, // LEFT-TO-RIGHT OVERRIDE
		0x202E: // RIGHT-TO-LEFT OVERRIDE

		return true
	}
	return false
}
