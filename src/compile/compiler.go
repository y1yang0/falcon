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
package compile

import (
	"falcon/ast"
	"falcon/compile/codegen"
	"falcon/compile/ssa"
	"falcon/utils"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const DebugPrintTypedAst = true
const DebugPrintAst = false
const DebugPrintLexicalToken = false
const DebugDumpAst = false
const DebugDumpSSA = false

func compileAsm(wd string, asm string, fileName string) {
	asmName := filepath.Join(wd, fileName+".s")
	ioutil.WriteFile(asmName, []byte(asm), 0644)

	osType := runtime.GOOS
	switch osType {
	case "windows":
		utils.ExecuteCmd(wd, "cmd.exe", "/c", "gcc", "-g", "-c", fileName+".s")
	case "darwin", "linux":
		utils.ExecuteCmd(wd, "gcc", "-g", "-c", fileName+".s")
	default:
		utils.Unimplement()
	}
}

func getLibNameFromPath(filePath string) string {
	// e.g. /path/to/filename.y -> filename
	filenameWithExt := filepath.Base(filePath)
	filename := strings.TrimSuffix(filenameWithExt, filepath.Ext(filenameWithExt))
	return filename
}

func parseY(filePath string, debug bool) ast.AstDecl {
	libName := getLibNameFromPath(filePath)
	if debug {
		if DebugPrintLexicalToken {
			fmt.Printf("== Lexing(%s) ==\n", filePath)
			ast.PrintTokenized(filePath)
		}
	}
	// parse the source file to untyped AST
	root := ast.ParseFile(filePath)
	if debug {
		if DebugPrintAst {
			fmt.Printf("== AST(%s) ==\n", filePath)
			ast.PrintAst(root, true)
		}
		// Dump ast when you want
		if DebugDumpAst {
			ast.DumpAstToDotFile(libName, root)
		}
	}
	return root
}

func compileY(wd string, filePath string, debug bool, root *ast.PackageDecl) {
	libName := getLibNameFromPath(filePath)

	lirs := make([]*codegen.LIR, 0)
	for _, funcDecl := range root.Func {
		decl := funcDecl.(*ast.FuncDecl)
		fn := ssa.Compile(decl, debug)
		if debug {
			if DebugDumpSSA {
				ssa.DumpSSAToDotFile(fn)
			}
		}
		lir := codegen.Lower(fn)
		if debug {
			fmt.Printf("== LIR(%s) ==\n", decl.Name)
			fmt.Printf("%s\n", lir)
		}
		lirs = append(lirs, lir)
	}

	// Generate assembly code
	text := codegen.CodeGen(lirs, debug)

	if debug {
		fmt.Printf("== ASM(%s.s) ==\n", filePath)
		fmt.Printf("%s\n", text)
	}

	compileAsm(wd, text, libName)
	fmt.Printf("Compiling %s to %s.o\n", filePath, libName)
}

func compileC(wd string, filePath string) {
	libName := getLibNameFromPath(filePath)
	osType := runtime.GOOS
	switch osType {
	case "windows":
		utils.ExecuteCmd(wd, "cmd.exe", "/c", "gcc", "-g", "-c", libName+".c")
	case "darwin", "linux":
		utils.ExecuteCmd(wd, "gcc", "-g", "-std=c99", "-c", libName+".c")
	default:
		utils.Unimplement()
	}
}

func linkFiles(wd string, target string, files ...string) {
	osType := runtime.GOOS
	switch osType {
	case "windows":
		args := []string{"cmd.exe", "/c", "gcc", "-Wl,--entry=entrypoint", "-o", target}
		args = append(args, files...)
		utils.ExecuteCmd(wd, args...)
	case "darwin", "linux":
		args := []string{"gcc", "-Wl,--entry=entrypoint", "-g", "-o", target}
		args = append(args, files...)
		utils.ExecuteCmd(wd, args...)
	default:
		utils.Unimplement()
	}
}

// COMPILE THE WORLD :0
func CompileTheWorld(wd string, source string) string {
	// Create temp dir and copy dependencies into it
	filesToCopy := []string{
		filepath.Join(wd, "../lib", "stdlib.y"),
		filepath.Join(wd, "../lib", "runtime.c"),
		filepath.Join(wd, "../lib", "builtin.c"),
		filepath.Join(wd, "../lib", "falcon.h"),
		source,
	}
	tempDir, err := utils.CopyFilesToTempDir("", filesToCopy)
	if err != nil {
		fmt.Println("Failed to copy files:", err)
		os.Exit(1)
	}
	fmt.Printf("Compilation workspace: %s\n", tempDir)
	// defer os.RemoveAll(tempDir)

	// Compile the compiler
	stdLib := filepath.Join(tempDir, "stdlib.y")
	userCode := filepath.Join(tempDir, filepath.Base(source))
	root1 := parseY(stdLib, false).(*ast.PackageDecl)
	root2 := parseY(userCode, true).(*ast.PackageDecl)
	// Type inference requires the whole-world AST
	ast.InferTypes(true /*debug*/, root1, root2)
	ast.TypeCheck(true /*debug*/, root1, root2)

	compileY(tempDir, userCode, true /*debug*/, root2)
	compileY(tempDir, stdLib, false /*debug*/, root1)
	// Compile the runtime written by C
	compileC(tempDir, filepath.Join(tempDir, "builtin.c"))
	compileC(tempDir, filepath.Join(tempDir, "runtime.c"))
	// Link them together
	libName := getLibNameFromPath(source)
	linkFiles(tempDir, libName, "stdlib.o", libName+".o", "runtime.o", "builtin.o")

	// Copy the binary to the current working directory
	target := filepath.Join(wd, libName)
	if runtime.GOOS == "windows" {
		libName = libName + ".exe"
		target = target + ".exe"
	}
	err = utils.CopyFile(filepath.Join(tempDir, libName), target)
	if err != nil {
		fmt.Printf("Failed to copy the binary: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Done!\n")
	return target
}

func CompileText(source string) string {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "falcon_")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}
	_, err = tmpFile.WriteString(source)
	if err != nil {
		log.Fatal("Failed to write to temporary file", err)
	}
	defer os.Remove(tmpFile.Name())
	wd, _ := os.Getwd()
	parent := filepath.Dir(wd)
	return CompileTheWorld(parent, tmpFile.Name())
}
