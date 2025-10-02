package main

import (
	"bytes"
	"os/exec"
	"time"

	"github.com/yarlson/tap"
)

func main() {
	tap.Intro("🚀 Stream Example")

	// Example 1: Manual writes
	st := tap.NewStream(tap.StreamOptions{ShowTimer: true})
	st.Start("Building project")
	st.WriteLine("step 1: fetch deps")
	time.Sleep(300 * time.Millisecond)
	st.WriteLine("step 2: compile")
	time.Sleep(300 * time.Millisecond)
	st.WriteLine("step 3: link")
	st.Stop("Done", 0)

	// Example 2: Pipe external command output
	cmd := exec.Command("bash", "-lc", "printf 'line 1\\nline 2\\nline 3\\n'")

	var buf bytes.Buffer

	cmd.Stdout = &buf
	_ = cmd.Run()

	st2 := tap.NewStream(tap.StreamOptions{ShowTimer: false})
	st2.Start("Streaming command output")
	st2.Pipe(&buf)
	st2.Stop("OK", 0)

	tap.Outro("End of stream example")
}
