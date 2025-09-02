package tap

import "fmt"

// OSC 9;4 helpers. We use ST (ESC \\) terminator for robustness.
const oscTerm = "\x1b\\"

func oscSpin(w Writer) {
	if w == nil {
		return
	}
	_, _ = w.Write([]byte("\x1b]9;4;3" + oscTerm))
}

func oscClear(w Writer) {
	if w == nil {
		return
	}
	_, _ = w.Write([]byte("\x1b]9;4;0" + oscTerm))
}

func oscSet(w Writer, pct int) {
	if w == nil {
		return
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	_, _ = fmt.Fprintf(w, "\x1b]9;4;1;%d%s", pct, oscTerm)
}
