package main

import (
	"fmt"
	"os"
)

/*
#include <assert.h>
#include <stdlib.h>

void call_getenv() {
	const char* value = getenv("doesnotexist");
	assert(value == NULL);
}
*/
import "C"

func main() {
	fmt.Println("setting new environment variables; will not crash on Mac OS X")

	go func() {
		for {
			C.call_getenv()
		}
	}()

	for i := 0; i < 100000; i++ {
		os.Setenv(fmt.Sprintf("var_%d", i), "value")
	}
}
