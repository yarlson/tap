package tap

import (
	"context"
	"strings"
	"testing"
	"time"
)

// suggestFn is a helper to convert a static list into a suggestion function.
func suggestFn(list []string) func(string) []string {
	return func(input string) []string {
		var out []string

		if input == "" {
			return list
		}

		low := strings.ToLower(input)
		for _, s := range list {
			if strings.Contains(strings.ToLower(s), low) {
				out = append(out, s)
			}
		}

		return out
	}
}

func TestAutocomplete_EnterReturnsTypedInput(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	SetTermIO(in, out)
	defer SetTermIO(nil, nil)

	resultCh := make(chan string, 1)

	go func() {
		res := Autocomplete(context.Background(), AutocompleteOptions{
			Message:    "Package name:",
			Suggest:    suggestFn([]string{"go", "python", "java"}),
			MaxResults: 5,
		})
		resultCh <- res
	}()

	// Wait for prompt initialization, then type "py" and press Enter
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("p", Key{Name: "p"})
	in.EmitKeypress("y", Key{Name: "y"})
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "py" {
		t.Fatalf("expected 'py', got %q", got)
	}
}

func TestAutocomplete_TabAcceptsSuggestion(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	SetTermIO(in, out)
	defer SetTermIO(nil, nil)

	resultCh := make(chan string, 1)

	go func() {
		res := Autocomplete(context.Background(), AutocompleteOptions{
			Message:    "Language:",
			Suggest:    suggestFn([]string{"go", "golang", "python"}),
			MaxResults: 5,
		})
		resultCh <- res
	}()

	// Wait for init; type "go", accept suggestion (first match) via Tab, then Enter
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("g", Key{Name: "g"})
	in.EmitKeypress("o", Key{Name: "o"})
	in.EmitKeypress("\t", Key{Name: "tab"})
	time.Sleep(10 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "go" {
		t.Fatalf("expected 'go' after tab acceptance, got %q", got)
	}
}

func TestAutocomplete_ArrowNavigationChangesSelection(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	SetTermIO(in, out)
	defer SetTermIO(nil, nil)

	resultCh := make(chan string, 1)

	go func() {
		res := Autocomplete(context.Background(), AutocompleteOptions{
			Message:    "Pick:",
			Suggest:    suggestFn([]string{"alpha", "beta", "gamma"}),
			MaxResults: 5,
		})
		resultCh <- res
	}()

	// Wait for init; type "a" (matches alpha, beta, gamma). Move down to select beta, accept via Tab, then Enter
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("a", Key{Name: "a"})
	in.EmitKeypress("", Key{Name: "down"})
	in.EmitKeypress("\t", Key{Name: "tab"})
	time.Sleep(10 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "beta" {
		t.Fatalf("expected 'beta', got %q", got)
	}
}
