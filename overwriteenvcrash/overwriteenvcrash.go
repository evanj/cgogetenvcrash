package main

import (
	"fmt"
	"os"
)

/*
#include <assert.h>
#include <stdlib.h>
#include <string.h>

void call_getenv() {
	const char* value = getenv("doesnotexist");
	assert(value == NULL);
}
*/
import "C"

const varName = "go_var"

func main() {
	fmt.Println("calling setenv to overwrite a single environment variable many times: does not crash")

	go func() {
		for {
			C.call_getenv()
		}
	}()

	for i := 0; i < 1000000; i++ {
		os.Setenv(varName, fmt.Sprintf("i->%d", i))
	}
}
