package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

/*
#include <assert.h>
#include <stdlib.h>
#include <string.h>

static const char VAR_NAME[] = "go_var";
static const char VAR_VALUE[] = "value";

void call_getenv_any_value() {
	const char* value = getenv(VAR_NAME);
	assert(value == NULL || strlen(value) > 0);
}

void call_getenv_expected_value() {
	const char* value = getenv(VAR_NAME);
	assert(value == NULL || strcmp(value, VAR_VALUE) == 0);
}

void call_getenv_expected_not_found() {
	const char* value = getenv(VAR_NAME);
	assert(value == NULL);
}
*/
import "C"

const varName = "go_var"
const varValue = "value"

const numTestIterations = 50000

func runTest(cFunc func(), goFunc func(iteration int)) {
	start := time.Now()
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}

			cFunc()
		}
	}()

	for i := 0; i < numTestIterations; i++ {
		goFunc(i)
	}
	fmt.Printf("   %d iterations completed in %s\n", numTestIterations, time.Since(start))
}

// testDefinition defines an environment variable test.
type testDefinition struct {
	// nameID uniquely identifies this test for the command line flag
	nameID string
	// description is printed as a human readable descriptio
	description string
	// cFunc is a Cgo function that calls getenv
	cFunc func()
	// goFunc is the Go function that changes the environment
	goFunc func(iteration int)
	// setupFunc is called before the test starts
	setupFunc func()
}

func main() {
	const runSingleTestFlag = "runSingleTest"
	runSingleTest := flag.String(runSingleTestFlag, "", "specific test to run")
	flag.Parse()

	// used for the one var / different length test
	bigString := ""

	doNothingSetupFunc := func() {}
	tests := []testDefinition{
		{
			"setenv_one_var_same_length",
			"Setenv overwriting one variable with different strings same length (no known crashes)",
			func() { C.call_getenv_any_value() },
			func(iteration int) {
				// buggy with musl but strings are short and the same length so may be reused
				os.Setenv(varName, fmt.Sprintf("i is %d", iteration))
			},
			doNothingSetupFunc,
		},
		{
			"setenv_one_var_different_length",
			"Setenv overwriting one variable with different strings different length (crashes musl)",
			func() { C.call_getenv_any_value() },
			func(iteration int) {
				// different sized strings cause free() to deallocate causing crash
				os.Setenv(varName, bigString[:iteration])
			},
			func() {
				// somewhat inefficient way to build a big string
				bigSBuilder := &strings.Builder{}
				for i := 0; i < numTestIterations; i++ {
					bigSBuilder.WriteByte('x')
				}
				bigString = bigSBuilder.String()
			},
		},
		{
			"set_unset_one_var",
			"Setenv/Unsetenv of one variable (crashes musl, Mac OS X)",
			func() { C.call_getenv_expected_value() },
			func(iteration int) {
				err := os.Setenv(varName, varValue)
				if err != nil {
					panic(err)
				}
				err = os.Unsetenv(varName)
				if err != nil {
					panic(err)
				}
			},
			doNothingSetupFunc,
		},
		{
			"many_new_vars",
			"Setenv new variables (crashes Linux glibc, musl)",
			func() { C.call_getenv_expected_not_found() },
			func(iteration int) {
				err := os.Setenv(fmt.Sprintf("%s_%d", varName, iteration), varValue)
				if err != nil {
					panic(err)
				}
			},
			doNothingSetupFunc,
		},
		{
			"unsetenv_many",
			"Unsetenv many variables (crashes musl)",
			func() { C.call_getenv_expected_not_found() },
			func(iteration int) {
				err := os.Unsetenv(fmt.Sprintf("%s_%d", varName, iteration))
				if err != nil {
					panic(err)
				}
			},
			func() {
				// create all the environment variables
				for i := 0; i < numTestIterations; i++ {
					err := os.Setenv(fmt.Sprintf("%s_%d", varName, i), varValue)
					if err != nil {
						panic(err)
					}
				}
			},
		},
	}

	// run a single test in this process
	if *runSingleTest != "" {
		// find the test and run it
		for _, test := range tests {
			if test.nameID == *runSingleTest {
				test.setupFunc()
				runTest(test.cFunc, test.goFunc)
				os.Exit(0)
			}
		}
		fmt.Fprintf(os.Stderr, "ERROR: could not find test with name %#v\n", *runSingleTest)
		os.Exit(1)
	}

	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	// run all tests in subprocesses
	for _, test := range tests {
		fmt.Printf("## %s ...\n", test.description)

		subproc := exec.Command(exePath, "--"+runSingleTestFlag, test.nameID)
		out, err := subproc.CombinedOutput()
		if err != nil {
			fmt.Println("--- FAILED TEST OUTPUT ---")
			os.Stdout.Write(out)
			fmt.Println("--- END FAILED OUTPUT ---")
		} else {
			fmt.Println("  PASSED (no crash)")
		}
	}
}
