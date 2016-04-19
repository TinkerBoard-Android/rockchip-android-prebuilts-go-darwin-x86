// Do not edit. Bootstrap copy of /Volumes/Android/buildbot/src/android/build-tools/out/obj/go/src/cmd/compile/internal/gc/mkbuiltin.go

//line /Volumes/Android/buildbot/src/android/build-tools/out/obj/go/src/cmd/compile/internal/gc/mkbuiltin.go:1
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// Generate builtin.go from $* (runtime.go and unsafe.go).
// Run this after changing runtime.go and unsafe.go
// or after changing the export metadata format in the compiler.
// Either way, you need to have a working compiler binary first.
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	f, err := os.Create("builtin.go")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	fmt.Fprintln(w, "// AUTO-GENERATED by mkbuiltin.go; DO NOT EDIT")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "package gc")

	for _, name := range os.Args[1:] {
		mkbuiltin(w, name)
	}

	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}
}

// Compile .go file, import data from .6 file, and write Go string version.
func mkbuiltin(w io.Writer, name string) {
	if err := exec.Command("go", "tool", "compile", "-A", "builtin/"+name+".go").Run(); err != nil {
		log.Fatal(err)
	}
	obj := name + ".o"
	defer os.Remove(obj)

	r, err := os.Open(obj)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	scanner := bufio.NewScanner(r)

	// Look for $$ that introduces imports.
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "$$") {
			goto Begin
		}
	}
	log.Fatal("did not find beginning of imports")

Begin:
	initfunc := fmt.Sprintf("init_%s_function", name)

	fmt.Fprintf(w, "\nconst %simport = \"\" +\n", name)

	// sys.go claims to be in package PACKAGE to avoid
	// conflicts during "go tool compile sys.go".  Rename PACKAGE to $2.
	replacer := strings.NewReplacer("PACKAGE", name)

	// Process imports, stopping at $$ that closes them.
	for scanner.Scan() {
		p := scanner.Text()
		if strings.Contains(p, "$$") {
			goto End
		}

		// Chop leading white space.
		p = strings.TrimLeft(p, " \t")

		// Cut out decl of init_$1_function - it doesn't exist.
		if strings.Contains(p, initfunc) {
			continue
		}

		fmt.Fprintf(w, "\t%q +\n", replacer.Replace(p)+"\n")
	}
	log.Fatal("did not find end of imports")

End:
	fmt.Fprintf(w, "\t\"$$\\n\"\n")
}
