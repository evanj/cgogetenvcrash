package main

import (
	"fmt"
	"os"
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
*/
import "C"

const varName = "go_var"
const varValue = "value"

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

	const numIterations = 1000
	for i := 0; i < numIterations; i++ {
		goFunc(i)
	}
	fmt.Printf("   %d iterations completed in %s\n", numIterations, time.Since(start))
}

func main() {
	fmt.Println("## Setenv overwriting a single variable with unique strings")
	fmt.Println("   does not crash Mac OS X or Linux glibc")
	runTest(func() { C.call_getenv_any_value() }, func(iteration int) {
		os.Setenv(varName, fmt.Sprintf("i->%d", iteration))
	})
	os.Unsetenv(varName)

	fmt.Println("## Setenv/Unsetenv of single variable")
	fmt.Println("   crashes Mac OS X, not Linux glibc")
	runTest(func() { C.call_getenv_expected_value() }, func(iteration int) {
		os.Setenv(varName, varValue)
		os.Unsetenv(varName)
	})
	os.Unsetenv(varName)

	fmt.Println("## Setenv of many new variables")
	fmt.Println("   crashes Linux glibc, not Mac OS X")
	runTest(func() { C.call_getenv_expected_value() }, func(iteration int) {
		os.Setenv(fmt.Sprintf("%s_%d", varName, iteration), varValue)
	})
}
