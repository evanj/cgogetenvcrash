package main

import (
	"fmt"
	"os"
)

/*
#include <stdlib.h>

void call_getenv() {

	const char* wtf = getenv("doesnotexist");
	*wtf;
}
*/
import "C"

func main() {
	go func() {
		for {
			C.call_getenv()
		}
	}()

	for i := 0; i < 100000; i++ {
		os.Setenv(fmt.Sprintf("var_%d", i), "value")
	}
}
