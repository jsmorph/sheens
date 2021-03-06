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

// Package main is a command-line program for spec testing.
package main

import (
	"context"
	"flag"
	"io/ioutil"
	"time"

	"github.com/jsmorph/sheens/interpreters"
	"github.com/jsmorph/sheens/tools/expect"

	"github.com/jsccast/yaml"
)

func main() {

	var (
		testFilename = flag.String("f", "specs/tests/double.test.yaml", "filename for test session")
		dir          = flag.String("d", ".", "working directory")
		showStderr   = flag.Bool("show-err", false, "show subprocess stderr")
		showStdin    = flag.Bool("show-in", false, "show subprocess stdin")
		showStdout   = flag.Bool("show-out", false, "show subprocess stdout")
		timeout      = flag.Duration("t", 10*time.Second, "main timeout")
	)

	flag.Parse()

	cmd := flag.Args()

	bs, err := ioutil.ReadFile(*testFilename)
	if err != nil {
		panic(err)
	}

	var s expect.Session
	if err = yaml.Unmarshal(bs, &s); err != nil {
		panic(err)
	}

	s.Interpreters = interpreters.Standard()
	s.ShowStdin = *showStdin
	s.ShowStdout = *showStdout
	s.ShowStderr = *showStderr

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if err = s.Run(ctx, *dir, cmd...); err != nil {
		panic(err)
	}
}
