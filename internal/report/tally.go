package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

const targetScore = 80

// Tally tracks lint section scores.
type Tally struct {
	counts []int
	score  int
	valid  bool
}

// NewTally returns a new tally.
func NewTally() *Tally {
	return &Tally{counts: make([]int, 4)}
}

// Score returns the tally computed score.
func (t *Tally) Score() int {
	return t.score
}

// IsValid checks if tally is valid.
func (t *Tally) IsValid() bool {
	return t.valid
}

// Rollup tallies up the report scores.
func (t *Tally) Rollup(o issues.Outcome) *Tally {
	if o == nil || len(o) == 0 {
		return t
	}

	t.valid = true
	for k := range o {
		t.counts[o.MaxSeverity(k)]++
	}
	t.computeScore()

	return t
}

// ComputeScore calculates the completed run score.
func (t *Tally) computeScore() int {
	var total, ok int
	for i, v := range t.counts {
		if i < 2 {
			ok += v
		}
		total += v
	}
	t.score = int(sanitize.ToPerc(int64(ok), int64(total)))

	return t.score
}

// Write out a tally.
func (t *Tally) write(w io.Writer, s *Sanitizer) {
	for i := len(t.counts) - 1; i >= 0; i-- {
		emoji := s.EmojiForLevel(issues.Level(i))
		fmat := "%s %d "
		if s.jurassicMode {
			fmat = "%s:%d "
		}
		fmt.Fprintf(w, fmat, emoji, t.counts[i])
	}

	score, color := t.score, ColorAqua
	if score < targetScore {
		color = ColorRed
	}
	percentageSign := "٪"
	if s.jurassicMode {
		percentageSign = "%%"
	}
	fmt.Fprintf(w, "%s%s", s.Color(strconv.Itoa(score), color), percentageSign)
}

// Dump writes out tally and computes length
func (t *Tally) Dump(s *Sanitizer) string {
	w := bytes.NewBufferString("")
	t.write(w, s)

	return w.String()
}

// MarshalYAML renders a tally to YAML.
func (t *Tally) MarshalYAML() (interface{}, error) {
	y := struct {
		OK    int `yaml:"ok"`
		Info  int `yaml:"info"`
		Warn  int `yaml:"warning"`
		Error int `yaml:"error"`
		Score int `yaml:"score"`
	}{
		Score: t.score,
	}

	for i, v := range t.counts {
		switch i {
		case 0:
			y.OK = v
		case 1:
			y.Info = v
		case 2:
			y.Warn = v
		case 3:
			y.Error = v
		}
	}

	return y, nil
}

// MarshalJSON renders a tally to JSON.
func (t *Tally) MarshalJSON() ([]byte, error) {
	y := struct {
		OK    int `json:"ok"`
		Info  int `json:"info"`
		Warn  int `json:"warning"`
		Error int `json:"error"`
		Score int `json:"score"`
	}{
		Score: t.score,
	}

	for i, v := range t.counts {
		switch i {
		case 0:
			y.OK = v
		case 1:
			y.Info = v
		case 2:
			y.Warn = v
		case 3:
			y.Error = v
		}
	}

	return json.Marshal(y)
}

// ----------------------------------------------------------------------------
// Helpers...

func toPerc(v1, v2 float64) float64 {
	if v2 == 0 {
		return 0
	}
	return (v1 / v2) * 100
}
