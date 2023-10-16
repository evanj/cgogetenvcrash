# Go netdns=cgo crash (SIGSEGV) in getaddrinfo

A Go program that uses the cgo DNS resolver can cause a segmentation fault if it sets environment variables while looking up DNS names. The problem is that glibc's DNS resolver can be configured using a number of environment variables ([see `man resolv.conf`](https://man7.org/linux/man-pages/man5/resolv.conf.5.html)). However, glibc's implementation of getenv and setenv are not thread-safe. This is permitted by the POSIX standard, and clearly documented in [`man setenv`](https://man7.org/linux/man-pages/man3/setenv.3.html) and [`man attributes`](https://man7.org/linux/man-pages/man7/attributes.7.html). This program does *not* crash using musl libc, because its DNS resolver does not use environment variables. However, a program that calls into C code that calls `getenv` will also crash, either with musl or glibc.


## Reproduction instructions:

1. Build the binary with `CGO_ENABLED=1`.
2. Run the binary many times in a loop. I have to run it in a Docker container to make it crash for some reason.

Example command lines:

```
# Outside docker container
CGO_ENABLED=1 go build -o cgocrashbug .
docker run --rm -ti --mount type=bind,source=$(pwd),destination=/cgogetenvcrash debian:latest

# Inside the docker container
for i in $(seq 10000); do GOTRACEBACK=crash GODEBUG=netdns=2+cgo /cgogetenvcrash/cgogetenvcrash || break; done
```

## Notes about Go's setenv implementation

* `os.Setenv`: Calls `syscall.Setenv` https://github.com/golang/go/blob/dc12cb179a3fb97bf9a12155c742f1737e858f7c/src/os/env.go#L119
* `syscall.Setenv`: Copies the process's initial environment variables into its own copy, protected by a Mutex. The mutex is held while calling all the functions below, including C's setenv() function, so only one Go thread can modify environment variables at a time. https://github.com/golang/go/blob/dc12cb179a3fb97bf9a12155c742f1737e858f7c/src/syscall/env_unix.go#L91
* `runtime.Setenv`: calls `setenv_c`, and if the key is GODEBUG, updates some internal state. https://github.com/golang/go/blob/dc12cb179a3fb97bf9a12155c742f1737e858f7c/src/runtime/runtime.go#L130
* `setenv_c`: Calls C's setenv() function if Cgo is enabled. https://github.com/golang/go/blob/dc12cb179a3fb97bf9a12155c742f1737e858f7c/src/runtime/env_posix.go#L49


## Raw crash output followed by gdb backtrace

```
build setting CGO_ENABLED="1"; (can crash)
go package net: confVal.netCgo = true  netGo = false
go package net: using cgo DNS resolver
go package net: hostLookupOrder(localhost) = cgo
SIGSEGV: segmentation violation
PC=0x7f5740e450cd m=0 sigcode=1
signal arrived during cgo execution

goroutine 20 [syscall]:
runtime.cgocall(0x401000, 0xc00003ed88)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/cgocall.go:157 +0x4b fp=0xc00003ed60 sp=0xc00003ed28 pc=0x40568b
net._C2func_getaddrinfo(0xc0000a6700, 0x0, 0xc00009e420, 0xc00009a040)
	_cgo_gotypes.go:100 +0x55 fp=0xc00003ed88 sp=0xc00003ed60 pc=0x4c6175
net._C_getaddrinfo.func1(0x40e3e5?, 0x8?, 0x4d4120?, 0x1?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/cgo_unix_cgo.go:78 +0x7a fp=0xc00003edf0 sp=0xc00003ed88 pc=0x4c653a
net._C_getaddrinfo(0x4faaeb?, 0x9?, 0x0?, 0x0?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/cgo_unix_cgo.go:78 +0x13 fp=0xc00003ee20 sp=0xc00003edf0 pc=0x4c6473
net.cgoLookupHostIP({0x4fa26d, 0x2}, {0x4faaeb, 0x9})
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/cgo_unix.go:166 +0x24f fp=0xc00003ef60 sp=0xc00003ee20 pc=0x4a9def
net.cgoLookupIP.func1()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/cgo_unix.go:215 +0x25 fp=0xc00003ef90 sp=0xc00003ef60 pc=0x4aa505
net.doBlockingWithCtx[...].func1()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/cgo_unix.go:56 +0x35 fp=0xc00003efe0 sp=0xc00003ef90 pc=0x4c67f5
runtime.goexit()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc00003efe8 sp=0xc00003efe0 pc=0x465ca1
created by net.doBlockingWithCtx[...] in goroutine 5
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/cgo_unix.go:54 +0xd8

goroutine 1 [runnable]:
runtime.evacuate_faststr(0x4ddc20, 0xc00009e390, 0x0?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/map_faststr.go:402 +0x3da fp=0xc000052d18 sp=0xc000052d10 pc=0x413eda
runtime.growWork_faststr(0xc000052da8?, 0xc00009e390, 0x5be830?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/map_faststr.go:398 +0x5f fp=0xc000052d48 sp=0xc000052d18 pc=0x413abf
runtime.mapassign_faststr(0x4ddc20, 0xc00009e390, {0xc000016110, 0xb})
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/map_faststr.go:227 +0x125 fp=0xc000052db8 sp=0xc000052d48 pc=0x413405
syscall.Setenv({0xc000016110, 0xb}, {0x4fa2ad, 0x3})
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/syscall/env_unix.go:121 +0x325 fp=0xc000052e58 sp=0xc000052db8 pc=0x47d365
os.Setenv({0xc000016110?, 0xc?}, {0x4fa2ad?, 0x1?})
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/os/env.go:120 +0x25 fp=0xc000052e98 sp=0xc000052e58 pc=0x48e9a5
main.main()
	/home/ej/cgogetenvcrash/cgogetenvcrash.go:44 +0x1fa fp=0xc000052f40 sp=0xc000052e98 pc=0x4cba9a
runtime.main()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:267 +0x2bb fp=0xc000052fe0 sp=0xc000052f40 pc=0x43895b
runtime.goexit()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc000052fe8 sp=0xc000052fe0 pc=0x465ca1

goroutine 2 [force gc (idle)]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:398 +0xce fp=0xc000042fa8 sp=0xc000042f88 pc=0x438dae
runtime.goparkunlock(...)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:404
runtime.forcegchelper()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:322 +0xb3 fp=0xc000042fe0 sp=0xc000042fa8 pc=0x438c33
runtime.goexit()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc000042fe8 sp=0xc000042fe0 pc=0x465ca1
created by runtime.init.6 in goroutine 1
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:310 +0x1a

goroutine 3 [GC sweep wait]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:398 +0xce fp=0xc000043778 sp=0xc000043758 pc=0x438dae
runtime.goparkunlock(...)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:404
runtime.bgsweep(0x0?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/mgcsweep.go:280 +0x94 fp=0xc0000437c8 sp=0xc000043778 pc=0x4250d4
runtime.gcenable.func1()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/mgc.go:200 +0x25 fp=0xc0000437e0 sp=0xc0000437c8 pc=0x41a465
runtime.goexit()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc0000437e8 sp=0xc0000437e0 pc=0x465ca1
created by runtime.gcenable in goroutine 1
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/mgc.go:200 +0x66

goroutine 4 [GC scavenge wait]:
runtime.gopark(0xc00006c000?, 0x521cb0?, 0x1?, 0x0?, 0xc0000091e0?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:398 +0xce fp=0xc000043f70 sp=0xc000043f50 pc=0x438dae
runtime.goparkunlock(...)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:404
runtime.(*scavengerState).park(0x5c73e0)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/mgcscavenge.go:425 +0x49 fp=0xc000043fa0 sp=0xc000043f70 pc=0x422969
runtime.bgscavenge(0x0?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/mgcscavenge.go:653 +0x3c fp=0xc000043fc8 sp=0xc000043fa0 pc=0x422efc
runtime.gcenable.func2()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/mgc.go:201 +0x25 fp=0xc000043fe0 sp=0xc000043fc8 pc=0x41a405
runtime.goexit()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc000043fe8 sp=0xc000043fe0 pc=0x465ca1
created by runtime.gcenable in goroutine 1
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/mgc.go:201 +0xa5

goroutine 18 [finalizer wait]:
runtime.gopark(0x4f8960?, 0x100439f01?, 0x0?, 0x0?, 0x440f65?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:398 +0xce fp=0xc000042628 sp=0xc000042608 pc=0x438dae
runtime.runfinq()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/mfinal.go:193 +0x107 fp=0xc0000427e0 sp=0xc000042628 pc=0x4194e7
runtime.goexit()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc0000427e8 sp=0xc0000427e0 pc=0x465ca1
created by runtime.createfing in goroutine 1
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/mfinal.go:163 +0x3d

goroutine 19 [select]:
runtime.gopark(0xc000057e90?, 0x2?, 0x8?, 0x31?, 0xc000057dec?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:398 +0xce fp=0xc000057c38 sp=0xc000057c18 pc=0x438dae
runtime.selectgo(0xc000057e90, 0xc000057de8, 0xc?, 0x0, 0x0?, 0x1)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/select.go:327 +0x725 fp=0xc000057d58 sp=0xc000057c38 pc=0x4487e5
net.(*Resolver).lookupIPAddr(0x5c7100, {0x523710?, 0x5f4c80}, {0x4fa26d, 0x2}, {0x4faaeb, 0x9})
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/lookup.go:332 +0x3fe fp=0xc000057f38 sp=0xc000057d58 pc=0x4bcf7e
net.(*Resolver).LookupIPAddr(...)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/lookup.go:210
net.LookupIP({0x4faaeb?, 0x0?})
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/lookup.go:196 +0x49 fp=0xc000057fb8 sp=0xc000057f38 pc=0x4bca09
main.main.func1()
	/home/ej/cgogetenvcrash/cgogetenvcrash.go:32 +0x28 fp=0xc000057fe0 sp=0xc000057fb8 pc=0x4cbb08
runtime.goexit()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc000057fe8 sp=0xc000057fe0 pc=0x465ca1
created by main.main in goroutine 1
	/home/ej/cgogetenvcrash/cgogetenvcrash.go:31 +0x1a6

goroutine 5 [select]:
runtime.gopark(0xc000058b50?, 0x2?, 0xa0?, 0x81?, 0xc000058b34?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:398 +0xce fp=0xc0000589e0 sp=0xc0000589c0 pc=0x438dae
runtime.selectgo(0xc000058b50, 0xc000058b30, 0x27?, 0x0, 0x437ad5?, 0x1)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/select.go:327 +0x725 fp=0xc000058b00 sp=0xc0000589e0 pc=0x4487e5
net.doBlockingWithCtx[...]({0x523780, 0xc000078050}, 0xc00009e3c0)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/cgo_unix.go:60 +0x14f fp=0xc000058bd0 sp=0xc000058b00 pc=0x4c7def
net.cgoLookupIP({0x523780, 0xc000078050}, {0x4fa26d, 0x2}, {0x4faaeb, 0x9})
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/cgo_unix.go:214 +0xb4 fp=0xc000058c00 sp=0xc000058bd0 pc=0x4aa474
net.(*Resolver).lookupIP(0x5c7100, {0x523780, 0xc000078050}, {0x4fa26d, 0x2}, {0x4faaeb, 0x9})
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/lookup_unix.go:70 +0x11a fp=0xc000058e58 sp=0xc000058c00 pc=0x4be3ba
net.(*Resolver).lookupIP-fm({0x523780?, 0xc000078050?}, {0x4fa26d?, 0x0?}, {0x4faaeb?, 0x0?})
	<autogenerated>:1 +0x49 fp=0xc000058ea0 sp=0xc000058e58 pc=0x4c9049
net.glob..func1({0x523780?, 0xc000078050?}, 0x0?, {0x4fa26d?, 0x0?}, {0x4faaeb?, 0x0?})
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/hook.go:23 +0x37 fp=0xc000058ee0 sp=0xc000058ea0 pc=0x4b6db7
net.(*Resolver).lookupIPAddr.func1()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/net/lookup.go:324 +0x3a fp=0xc000058f38 sp=0xc000058ee0 pc=0x4bd99a
internal/singleflight.(*Group).doCall(0x5c7110, 0xc0000780a0, {0xc000016060, 0xc}, 0xc00008e0c0?)
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/internal/singleflight/singleflight.go:93 +0x35 fp=0xc000058fa8 sp=0xc000058f38 pc=0x4a7975
internal/singleflight.(*Group).DoChan.func1()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/internal/singleflight/singleflight.go:86 +0x30 fp=0xc000058fe0 sp=0xc000058fa8 pc=0x4a7910
runtime.goexit()
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc000058fe8 sp=0xc000058fe0 pc=0x465ca1
created by internal/singleflight.(*Group).DoChan in goroutine 19
	/home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/internal/singleflight/singleflight.go:86 +0x2e9

rax    0xb
rbx    0x19ac
rcx    0x0
rdx    0x7f5740e0371c
rdi    0x7f5740f9f87b
rsi    0x7ffe40eea560
rbp    0x19ac090
rsp    0x7ffe40eea3a0
r8     0x0
r9     0x7ffe40eea430
r10    0xc0000a6700
r11    0xb
r12    0x4f4c
r13    0x7f5740f9f87d
r14    0xb
r15    0x9
rip    0x7f5740e450cd
rflags 0x10206
cs     0x33
fs     0x0
gs     0x0
```

## gdb stack trace

```
(gdb) thread 2
[Switching to thread 2 (Thread 0x7f5740e03740 (LWP 153059))]
#0  runtime.usleep () at /home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/sys_linux_amd64.s:135
135	in /home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/sys_linux_amd64.s
(gdb) bt
#0  runtime.usleep () at /home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/sys_linux_amd64.s:135
#1  0x000000000044b9aa in runtime.sighandler (sig=11, info=<optimized out>, ctxt=<optimized out>, gp=<optimized out>)
    at /home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/signal_unix.go:769
#2  0x000000000044b02e in runtime.sigtrampgo (sig=11, info=0xc0000114b0, ctx=0xc000011380)
    at /home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/signal_unix.go:490
#3  0x00000000004677c6 in runtime.sigtramp () at /home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/sys_linux_amd64.s:352
#4  <signal handler called>
#5  __GI_getenv (name=0x7f5740f9f87d "CALDOMAIN", name@entry=0x7f5740f9f87b "LOCALDOMAIN") at ./stdlib/getenv.c:84
#6  0x00007f5740f5016c in __nscd_getai (key=key@entry=0xc0000a6700 "localhost", result=result@entry=0x7ffe40eea560, h_errnop=0x7f5740e0371c) at ./nscd/nscd_getai.c:47
#7  0x00007f5740ef7312 in get_nscd_addresses (res=0x7ffe40eea540, req=0xc00009e420, name=0xc0000a6700 "localhost") at ../sysdeps/posix/getaddrinfo.c:495
#8  gaih_inet (tmpbuf=0x7ffe40eea690, naddrs=<synthetic pointer>, pai=0x7ffe40eea510, req=0xc00009e420, service=<optimized out>, name=0xc0000a6700 "localhost")
    at ../sysdeps/posix/getaddrinfo.c:1173
#9  __GI_getaddrinfo (name=<optimized out>, service=<optimized out>, hints=0xc00009e420, pai=0xc00009a040) at ../sysdeps/posix/getaddrinfo.c:2398
#10 0x0000000000401034 in net(.text) ()
#11 0x000000c00003ed88 in ?? ()
#12 0x000000c000082b60 in ?? ()
#13 0x0000000000000008 in ?? ()
#14 0x000000c00003ed18 in ?? ()
#15 0x0000000000465928 in runtime.asmcgocall () at /home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/asm_amd64.s:872
#16 0x000000c0000081a0 in ?? ()
#17 0x00007ffe40eeac58 in ?? ()
#18 0x0000000000441128 in runtime.newproc.func1 () at /home/ej/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.21.3.linux-amd64/src/runtime/proc.go:4487
#19 0x000000c0000096c0 in ?? ()
#20 0x000000000046835f in runtime.newproc (fn=0x0) at <autogenerated>:1
#21 0x00000000005c7480 in runtime[scavenger] ()
#22 0x0000000000000000 in ?? ()
```



## Build system details

```
$ cat /etc/lsb-release 
DISTRIB_ID=Ubuntu
DISTRIB_RELEASE=20.04
DISTRIB_CODENAME=focal
DISTRIB_DESCRIPTION="Ubuntu 20.04.6 LTS"
$ go version
go version go1.21.3 linux/amd64
```
