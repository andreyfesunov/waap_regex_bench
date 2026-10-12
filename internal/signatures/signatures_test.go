package signatures

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	input := "# comment\nProbe TCP NULL q||\nmatch http m|^foo\\|bar$|is p/product/\nrarity 1\nsoftmatch http m/^a\\/b$/\nProbe UDP DNS q|x|\nmatch dns m@^x$@\n"
	sigs, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) != 3 {
		t.Fatalf("got %d signatures", len(sigs))
	}
	if sigs[0].Expression != `^foo\|bar$` || sigs[0].Flags != "is" || sigs[0].Line != 3 {
		t.Fatalf("wrong first signature: %+v", sigs[0])
	}
	if sigs[1].Expression != "^a/b$" || sigs[1].Kind != "softmatch" || sigs[1].Probe != "NULL" {
		t.Fatalf("wrong second signature: %+v", sigs[1])
	}
	if sigs[2].ID != 2 || sigs[2].ProbeOrder != 1 || sigs[2].Probe != "DNS" {
		t.Fatalf("wrong third signature: %+v", sigs[2])
	}
}

func TestMalformed(t *testing.T) {
	for _, input := range []string{"match http m|x|", "Probe TCP NULL q||\nmatch http m|unterminated", "Probe TCP"} {
		if _, err := Parse(strings.NewReader(input)); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}
