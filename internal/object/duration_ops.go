// langur/object/duration_ops.go

package object

func (l *Duration) Add(d2 Object) Object {
	switch r := d2.(type) {
	case *Duration:
		return &Duration{
			Years:       l.Years + r.Years,
			Months:      l.Months + r.Months,
			Days:        l.Days + r.Days,
			Hours:       l.Hours + r.Hours,
			Minutes:     l.Minutes + r.Minutes,
			Seconds:     l.Seconds + r.Seconds,
			Nanoseconds: l.Nanoseconds + r.Nanoseconds,
		}

	default:
		return nil
	}
}

func (l *Duration) Multiply(d2 Object) Object {
	switch r := d2.(type) {
	case *Boolean:
		if r.Value {
			return l
		}
		return ZeroDuration

	default:
		return nil
	}
}
