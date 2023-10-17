package main

import (
	"fmt"
	"os"
)

/*
#include <assert.h>
#include <stdlib.h>
#include <string.h>

static const char VAR_NAME[] = "go_var";
static const char VAR_VALUE[] = "value";

void call_getenv() {
	const char* value = getenv(VAR_NAME);
	assert(value == NULL || strcmp(value, VAR_VALUE) == 0);
}
*/
import "C"

const varName = "go_var"
const varValue = "value"

func main() {
	fmt.Println("calling setenv/unsetenv: crashes Mac OS X, not Linux glibc")

	go func() {
		for {
			C.call_getenv()
		}
	}()

	for i := 0; i < 100000; i++ {
		os.Setenv(varName, varValue)
		os.Unsetenv(varName)
	}
}
