// langur/str/io.go

package str

import (
	"fmt"
	"os"
)

func PrintErr(s string) {
	fmt.Fprint(os.Stderr, s)
}
func PrintLnErr(s string) {
	PrintErr(s + "\n")
}
