// Copyright (c) 2024 The Falcon Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"falcon/compile"
	"fmt"
	"os"
)

func main() {
	// if runtime.GOARCH != "amd64" {
	// utils.Unimplement()
	// }
	if len(os.Args) != 2 {
		fmt.Println("Usage: falcon test.y")
		os.Exit(1)
	}
	source := os.Args[1]
	wd, _ := os.Getwd()
	compile.CompileTheWorld(wd, source)
}
