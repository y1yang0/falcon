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
package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
)

func Assert(cond bool, format string, msg ...interface{}) {
	if !cond {
		panic(fmt.Sprintf(format, msg...))
	}
}

func Any[T comparable](c T, cs ...T) bool {
	for _, cc := range cs {
		if c == cc {
			return true
		}
	}
	return false
}

func Unimplement() {
	panic("Not implement yet")
}

func ShouldNotReachHere() {
	panic("Should not reach here")
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	println(msg)
	panic(msg)
}

func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func ExecuteCmd(workDir string, args ...string) string {
	if !CommandExists(args[0]) {
		fmt.Printf("Warning: Can not find %v\n", args[0])
	}
	cmd := exec.Command(args[0], args[1:]...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = workDir

	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		fmt.Printf("cmd.Run: %s failed: %s\n", err, err)
		fmt.Printf("out:\n%s\nerr:\n%scmd:%v\n\n", outStr, errStr, args)
		os.Exit(1)
	}
	return outStr
}

func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceFileStat.Mode())
}

func CopyFilesToTempDir(dir string, files []string) (string, error) {
	tempDir, err := ioutil.TempDir("", dir)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		destFile := filepath.Join(tempDir, filepath.Base(file))
		// fmt.Printf("Copying %s to %s\n", file, destFile)
		err := CopyFile(file, destFile)
		if err != nil {
			os.RemoveAll(tempDir)
			return "", err
		}
	}
	return tempDir, nil
}

func Align16(n int) int {
	return (n + 15) &^ 15
}

func Float64ToHex(f float64) string {
	hex := fmt.Sprintf("%x", math.Float64bits(f))
	return fmt.Sprintf("0x%s", hex)
}
