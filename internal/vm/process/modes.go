// langur/vm/process/modes.go

package process

import (
	"fmt"
	"langur/modes"
	"langur/object"
	"langur/str"
	"math"
	"os"
)

func (pr *Process) setMode(code int, setting object.Object) error {
	switch code {
	case modes.MODE_DIVISION_MAX_SCALE:
		i, ok := object.NumberToInt(setting)
		if !ok {
			return fmt.Errorf("Expected integer for mode division max scale, not %s", setting.TypeString())
		}
		if i < 0 || i > math.MaxInt32-2 {
			return fmt.Errorf("Integer %d for mode division max scale out of range", i)
		}
		pr.Modes.DivisionMaxScale = i
		// FIXME: not safe for concurrency
		return object.SetDivisionMaxScaleMode(i)

	case modes.MODE_ROUNDING:
		i, ok := object.NumberToInt(setting)
		if !ok {
			return fmt.Errorf("Expected rounding mode (from %s hash), not %s", modes.RoundHashName, setting.TypeString())
		}
		// ensure it's in the enumeration
		_, ok = modes.RoundHashModeNames[i]
		if !ok {
			return fmt.Errorf("Unknown rounding mode")
		}
		pr.Modes.Rounding = i
		// FIXME: not safe for concurrency
		modes.RoundingMode = i

	case modes.MODE_CONSOLE_TEXT_MODE:
		b, ok := setting.(*object.Boolean)
		if !ok {
			return fmt.Errorf("Expected Boolean for mode console text mode")
		}
		pr.Modes.ConsoleTextMode = b.Value

	case modes.MODE_NEW_FILE_PERMISSIONS:
		i, ok := object.NumberToInt(setting)
		if !ok {
			return fmt.Errorf("Expected integer for mode new file permissions, not %s", setting.TypeString())
		}
		if i < 0 || i > 0o777 {
			return fmt.Errorf("Integer 8x%s for mode new file permissions out of range", str.IntToStr(i, 8))
		}
		pr.Modes.NewFilePermissions = os.FileMode(i)

	case modes.MODE_NOW_INCLUDES_NANO:
		b, ok := setting.(*object.Boolean)
		if !ok {
			return fmt.Errorf("Expected Boolean for mode now includes nanoseconds")
		}
		pr.Modes.NowIncludesNano = b.Value

	default:
		bug("setMode", fmt.Sprintf("Unknown mode setting %d", code))
		return fmt.Errorf("Unknown mode setting %d", code)
	}

	return nil
}
