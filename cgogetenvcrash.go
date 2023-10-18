package main

import (
	"fmt"
	"net"
	"os"
	"runtime/debug"
)

func main() {
	// not required to reproduce the bug but makes it easier to debug when testing
	const cgoEnabledKey = "CGO_ENABLED"
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		panic("debug.ReadBuildInfo() should not fail")
	}
	for _, setting := range buildInfo.Settings {
		if setting.Key == cgoEnabledKey {
			detailString := "can crash"
			if setting.Value != "1" {
				detailString = "CANNOT CRASH: bug requires CGO"
			}
			fmt.Printf("build setting %s=%#v; (%s)\n", cgoEnabledKey, setting.Value, detailString)
			break
		}
	}

	addrsDone := make(chan struct{})
	go func() {
		addrs, err := net.LookupIP("localhost")
		if err != nil {
			panic(err)
		}
		if len(addrs) == 0 {
			panic("no addrs for localhost")
		}
		close(addrsDone)
	}()

	// setting many environment variables makes the bug much more likely
	for i := 0; i < 100; i++ {
		err := os.Setenv(fmt.Sprintf("ENV_VAR_%03d", i), "foo")
		if err != nil {
			panic(err)
		}
	}

	<-addrsDone
	fmt.Println("exiting successfully")
}
