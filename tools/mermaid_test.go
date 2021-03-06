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

package tools

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/jsmorph/sheens/core"
	"github.com/jsmorph/sheens/interpreters"

	"github.com/jsccast/yaml"
)

func TestMermaid(t *testing.T) {
	var (
		leaveFile = false
		filename  = "g.mermaid"
		// specFilename = "../specs/homeassistant.yaml"
		specFilename = ""
	)

	out, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}

	if !leaveFile {
		defer func() {
			log.Printf("removing %s", filename)
			if err := os.Remove(filename); err != nil {
				t.Fatal(err)
			}
		}()
	}

	var spec *core.Spec

	if specFilename == "" {
		if spec, err = core.TurnstileSpec(context.Background()); err != nil {
			t.Fatal(err)
		}
	} else {
		interpreters := interpreters.Standard()

		specSrc, err := ioutil.ReadFile(specFilename)
		if err != nil {
			t.Fatal(err)
		}
		if err = yaml.Unmarshal(specSrc, &spec); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		if err = spec.Compile(ctx, interpreters, true); err != nil {
			t.Fatal(err)
		}
	}

	if err := Mermaid(spec, out, nil, "", ""); err != nil {
		t.Fatal(err)
	}

}
