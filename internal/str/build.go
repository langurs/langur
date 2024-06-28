// langur/str/build.go

package str

import (
	"fmt"
	// "langur/cpoint"
	"strings"
)

// NOTE: This might check for surrogate code points, to fix them or generate errors.
func BuildString(parts []interface{}) (string, error) {
	var sb strings.Builder
	// var err error

	for i, p := range parts {
		switch p := p.(type) {
		case rune:
			// if cpoint.IsSurrogate(p) {
			// 	return "", fmt.Errorf("Error building string; independent surrogate code (%s) in section %d", cpoint.Display(p), i+1)
			// }
			sb.WriteRune(p)

		case []rune:
			// p, err = cpoint.ReplaceSurrogates(p)
			// if err != nil {
			// 	return "", fmt.Errorf("Error building string section %d; %s", i+1, err.Error())
			// }
			sb.WriteString(string(p))

		case string:
			sb.WriteString(p)

		case byte:
			if p > 127 {
				return "", fmt.Errorf("Error building string; section %d contains independent byte (%d) >127", i+1, p)
			}
			sb.WriteByte(p)

		case []byte:
			// TODO: check for and convert surrogate codes, or generate error for bad ones
			sb.Write(p)

		default:
			return "", fmt.Errorf("Error building string; section %d not of usable type", i+1)
		}
	}
	return sb.String(), nil
}
