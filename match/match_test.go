/* Copyright 2018 Comcast Cable Communications Management, LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/* This file might have changed after the fork. */

package match

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	. "github.com/jsmorph/sheens/util/testutil"
)

func TestExtendBindings(t *testing.T) {
	bs := NewBindings().Extend("likes", "queso")
	queso, _ := bs["likes"]
	if queso != "queso" {
		t.Fatal(queso)
	}
}

func TestExtendmBindings(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		bs, err := NewBindings().Extendm("likes", "tacos", "needs", 3)
		if err != nil {
			t.Fatal(err)
		}
		tacos, _ := bs["likes"]
		if tacos != "tacos" {
			t.Fatal(tacos)
		}
		needs, _ := bs["needs"]
		if needs != 3 {
			t.Fatal(needs)
		}
	})

	t.Run("nonStringKey", func(t *testing.T) {
		bs, err := NewBindings().Extendm("likes", "tacos", true, 3)
		if err == nil {
			t.Fatal(bs)
		}
	})

	t.Run("oddArgs", func(t *testing.T) {
		bs, err := NewBindings().Extendm("likes", "tacos", "needs", 3, "nope")
		if err == nil {
			t.Fatal(bs)
		}
	})

}

func TestRemoveBindings(t *testing.T) {
	bs, err := NewBindings().Extendm("likes", "tacos", "needs", 3)
	if err != nil {
		t.Fatal(err)
	}
	bs = bs.Remove("needs")
	if needs, have := bs["needs"]; have {
		t.Fatal(needs)
	}
}

func TestDeleteExcept(t *testing.T) {
	bs, err := NewBindings().Extendm("likes", "tacos", "needs", 3, "when", "now")
	if err != nil {
		t.Fatal(err)
	}
	bs = bs.DeleteExcept("likes")
	if needs, have := bs["needs"]; have {
		t.Fatal(needs)
	}
	if when, have := bs["when"]; have {
		t.Fatal(when)
	}
	if likes, _ := bs["likes"]; likes != "tacos" {
		t.Fatal(likes)
	}
}

func TestFudge(t *testing.T) {
	pairs := []struct {
		X interface{}
		Y float64
	}{
		{
			X: 1,
			Y: 1,
		},
		{
			X: int32(1),
			Y: 1,
		},
		{
			X: int64(1),
			Y: 1,
		},
		{
			X: float32(1),
			Y: 1,
		},
		{
			X: float64(1),
			Y: 1,
		},
	}
	for _, pair := range pairs {
		y := fudge(pair.X)
		if y != pair.Y {
			t.Fatal(pair, y)
		}
	}
}

type MatchTest struct {
	Pattern       interface{}              `json:"p"`
	Message       interface{}              `json:"m"`
	Bindings      map[string]interface{}   `json:"b,omitempty"`
	Expected      []map[string]interface{} `json:"w,omitempty"`
	Error         bool                     `json:"err,omitempty"`
	Title         string                   `json:"title,omitempty"`
	Doc           string                   `json:"doc,omitempty"`
	NoDoc         bool                     `json:"noDoc,omitempty"`
	Verbose       bool                     `json:"verbose,omitempty"`
	BenchmarkOnly bool                     `json:"benchmarkOnly,omitempty"`
}

func (t MatchTest) Name(i int) string {
	if t.Title == "" {
		return fmt.Sprintf("%d", i)
	} else {
		return fmt.Sprintf("%03d %s", i, t.Title)
	}
}

// Hmm.  According to the encoding.json docs:
//
// String values encode as JSON strings coerced to valid UTF-8,
// replacing invalid bytes with the Unicode replacement rune. The
// angle brackets "<" and ">" are escaped to "\u003c" and "\u003e" to
// keep some browsers from misinterpreting JSON output as
// HTML. Ampersand "&" is also escaped to "\u0026" for the same
// reason.

func JSON(x interface{}) string {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(x); err != nil {
		panic(err)
	}
	return buf.String()
}

func (t MatchTest) Fprintf(w io.Writer, i int) {
	i++
	title := t.Title
	if title == "" {
		title = "Anonymous example"
	}
	fmt.Fprintf(w, "\n## %d. %s\n\n", i, title)
	if t.Doc != "" {
		fmt.Fprintf(w, "\n%s\n", t.Doc)
	}
	fmt.Fprintf(w, "The pattern\n```JSON\n%s\n```\n\n", JSON(t.Pattern))
	fmt.Fprintf(w, "matched against\n```JSON\n%s\n```\n\n", JSON(t.Message))
	if t.Bindings != nil {
		fmt.Fprintf(w, "with bindings\n```JSON\n%s\n```\n\n", JSON(t.Bindings))
	}
	if t.Error {
		fmt.Fprintf(w, "should return an error.\n")
	} else {
		fmt.Fprintf(w, "should return\n```JSON\n%s\n```\n", JSON(t.Expected))
	}
}

func compareMatchResult(bss []Bindings, expected []map[string]interface{}) bool {
	if len(bss) != len(expected) {
		return false
	}

	m := make(map[int]map[string]interface{})
	for i, got := range bss {
		m[i] = map[string]interface{}(got)
	}

	for _, e := range expected {
		found := false
		for k, v := range m {
			if reflect.DeepEqual(e, v) {
				delete(m, k)
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return 0 == len(m)
}

func (mt *MatchTest) Run(t *testing.T, check bool) {
	bs := mt.Bindings
	if bs == nil {
		bs = make(Bindings)
	}
	bss, err := DefaultMatcher.Match(mt.Pattern, mt.Message, bs)
	if !check {
		return
	}
	if err != nil {
		if !mt.Error {
			t.Fatal(err)
		}
	} else {
		if mt.Error {
			t.Fatal("expected an error")
		}
	}

	if !compareMatchResult(bss, mt.Expected) {
		t.Fatalf("match test failed: bindings: %s pattern: %s message: %s got: %s expected: %s\n",
			JS(mt.Bindings), JS(mt.Pattern), JS(mt.Message), JS(bss), JS(mt.Expected))
	}
}

func getMatchTests() ([]MatchTest, error) {
	js, err := ioutil.ReadFile("match_test.json")
	if err != nil {
		return nil, err
	}
	var tests []MatchTest
	if err = json.Unmarshal(js, &tests); err != nil {
		return nil, err
	}
	return tests, nil
}

func TestMatch(t *testing.T) {
	tests, err := getMatchTests()
	if err != nil {
		t.Fatal(err)
	}
	md, err := os.Create("match.md")
	if err != nil {
		t.Fatal(err)
	}
	defer md.Close()

	fmt.Fprintf(md, `# Pattern matching examples

Generated from test cases.

`)

	for i, test := range tests {
		if test.BenchmarkOnly {
			continue
		}
		if !test.NoDoc {
			test.Fprintf(md, i)
		}
		t.Run(test.Name(i), func(t *testing.T) {
			test.Run(t, true)
		})
	}
}

func BenchmarkMatch(b *testing.B) {
	tests, err := getMatchTests()
	if err != nil {
		b.Fatal(err)
	}
	for i, test := range tests {
		b.Run(test.Name(i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				test.Run(nil, false)
			}
		})
	}
}

func TestMatchFunctionBasic(t *testing.T) {
	bss, err := Match(true, true, NewBindings())
	check := func() {
		if err != nil {
			t.Fatal(err)
		}
		if len(bss) != 1 {
			t.Fatal(bss)
		}
		if len(bss[0]) != 0 {
			t.Fatal(bss[0])
		}
	}
	check()

	m := DefaultMatcher
	bss, err = m.Matches(true, true)
	check()
}

func TestMatchUnknown(t *testing.T) {
	alien := struct{}{}
	t.Run("pat", func(t *testing.T) {
		bss, err := Match(alien, true, NewBindings())
		if err == nil {
			t.Fatal(bss)
		}
		if upt, is := err.(*UnknownPatternType); !is {
			t.Fatal(err)
		} else {
			upt.Error() // Coverage!
		}
	})

	t.Run("msg", func(t *testing.T) {
		bss, err := Match(true, alien, NewBindings())
		if err != nil {
			t.Fatal(bss)
		}
		if len(bss) != 0 {
			t.Fatal(bss)
		}
	})

}
