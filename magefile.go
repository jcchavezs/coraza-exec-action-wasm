// Copyright 2022 Juan Pablo Tosso and the OWASP Coraza contributors
// SPDX-License-Identifier: Apache-2.0

//go:build mage
// +build mage

package main

import (
	"errors"
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var golangCILintVer = "v1.48.0" // https://github.com/golangci/golangci-lint/releases
var gosImportsVer = "v0.1.5"    // https://github.com/rinchsan/gosimports/releases/tag/v0.1.5

var errRunGoModTidy = errors.New("go.mod/sum not formatted, commit changes")

// Format formats code in this repository.
func Format() error {
	if err := sh.RunV("go", "generate", "./..."); err != nil {
		return err
	}

	if err := sh.RunV("go", "mod", "tidy"); err != nil {
		return err
	}

	return sh.RunV("go", "run", fmt.Sprintf("github.com/rinchsan/gosimports/cmd/gosimports@%s", gosImportsVer),
		"-w",
		"-local",
		"github.com/corazawaf/coraza",
		".")
}

// Lint verifies code quality.
func Lint() error {
	if err := sh.RunV("go", "run", fmt.Sprintf("github.com/golangci/golangci-lint/cmd/golangci-lint@%s", golangCILintVer), "run"); err != nil {
		return err
	}

	if err := sh.RunV("go", "mod", "tidy"); err != nil {
		return err
	}

	if sh.Run("git", "diff", "--exit-code", "go.mod", "go.sum") != nil {
		return errRunGoModTidy
	}

	return nil
}

func Generate() error {
	if err := sh.RunV("go", "generate", "./..."); err != nil {
		return err
	}

	return sh.RunV("wasm2wat", "./testdata/hello-world.wasm", "--output=./testdata/hello-world.wat")
}

// Test runs all tests.
func Test() error {
	mg.SerialDeps(Generate)
	if err := sh.RunV("go", "test", "./..."); err != nil {
		return err
	}

	return nil
}
