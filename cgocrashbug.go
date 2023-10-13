package main

import (
	"net"
	"os"
)

func main() {
	// logger, err := zap.NewProduction()
	// if err != nil {
	// 	t.Fatal(err)
	// }

	os.Setenv("SOME_NEW_ENV_VAR", "some_env_var_value")

	// this *might* have made the bug easier to reproduce?
	// this might also be a red herring: the bug happens without it
	// crashes seem to happen if "go package net: hostLookupOrder(localhost) = cgo"
	// is printed before the log message
	lookupGoroutineStarted := make(chan struct{})

	addrsDone := make(chan struct{})
	go func() {
		close(lookupGoroutineStarted)

		addrs, err := net.LookupIP("localhost")
		if err != nil {
			panic(err)
		}
		if len(addrs) == 0 {
			panic("no addrs for localhost")
		}
		close(addrsDone)
	}()

	<-lookupGoroutineStarted
	// logger.Info("some log message")
	os.Stdout.WriteString("some log message\n")

	os.Setenv("OTHER_ENV_VAR", "foo")
	<-addrsDone
	// fmt.Println("net.LookupIP completed in other goroutine")
}
