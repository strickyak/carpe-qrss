// +build main

// Thanks to https://github.com/nf/giffy
// Modified 2018 by Henry Strickland (github.com/strickyak) to use flag and command line arguments, according to the following license.

/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This command giffy reads all the JPEG and PNG files from the command line arguments
// and writes them to an animated GIF as the file named by the -o flag.
package main

import (
	"flag"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/strickyak/carpe-qrss"
)

var OUT = flag.String("o", "/dev/stdout", "The output gif filename")
var DELAY = flag.Duration("delay", 100*time.Millisecond, "delay per frame")

func main() {
	flag.Parse()
	carpe.BuildAnimatedGif(flag.Args(), *DELAY, *OUT)
}
