# Go netdns=cgo crash

Reproduction instructions:

```
go build -o cgocrashbug .
docker run --rm -ti --mount type=bind,source=/home/ej/dd/cgocrashbug,destination=/cgocrashbug debian:12.2

cd /cgocrashbug

for i in $(seq 10000); do GODEBUG=netdns=2+cgo ./cgocrashbug || break; done
```

Example output showing one successful run then a crash:

```
go package net: confVal.netCgo = true  netGo = false
go package net: using cgo DNS resolver
go package net: hostLookupOrder(localhost) = cgo
some log message
SIGSEGV: segmentation violation
PC=0x7f9b8aa5e0cd m=0 sigcode=128
signal arrived during cgo execution

goroutine 8 [syscall]:
runtime.cgocall(0x401000, 0xc000052d88)
	/go/src/runtime/cgocall.go:157 +0x4b fp=0xc000052d60 sp=0xc000052d28 pc=0x40490b
net._C2func_getaddrinfo(0xc0000160d0, 0x0, 0xc000094450, 0xc000054048)
	_cgo_gotypes.go:100 +0x55 fp=0xc000052d88 sp=0xc000052d60 pc=0x4ab4f5
net._C_getaddrinfo.func1(0x40d665?, 0x8?, 0x4b6180?, 0x1?)
	/go/src/net/cgo_unix_cgo.go:78 +0x7a fp=0xc000052df0 sp=0xc000052d88 pc=0x4ab8ba
net._C_getaddrinfo(0x4d82b6?, 0x9?, 0x0?, 0x0?)
	/go/src/net/cgo_unix_cgo.go:78 +0x13 fp=0xc000052e20 sp=0xc000052df0 pc=0x4ab7f3
net.cgoLookupHostIP({0x4d7ae4, 0x2}, {0x4d82b6, 0x9})
	/go/src/net/cgo_unix.go:166 +0x24f fp=0xc000052f60 sp=0xc000052e20 pc=0x48f16f
net.cgoLookupIP.func1()
	/go/src/net/cgo_unix.go:215 +0x25 fp=0xc000052f90 sp=0xc000052f60 pc=0x48f885
net.doBlockingWithCtx[...].func1()
	/go/src/net/cgo_unix.go:56 +0x35 fp=0xc000052fe0 sp=0xc000052f90 pc=0x4abb75
runtime.goexit()
	/go/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc000052fe8 sp=0xc000052fe0 pc=0x464481
created by net.doBlockingWithCtx[...] in goroutine 7
	/go/src/net/cgo_unix.go:54 +0xd8

goroutine 1 [chan receive]:
runtime.gopark(0x0?, 0x0?, 0xa?, 0x0?, 0x48a520?)
	/go/src/runtime/proc.go:398 +0xce fp=0xc000065e60 sp=0xc000065e40 pc=0x437cee
runtime.chanrecv(0xc00008a120, 0x0, 0x1)
	/go/src/runtime/chan.go:583 +0x3cd fp=0xc000065ed8 sp=0xc000065e60 pc=0x406ccd
runtime.chanrecv1(0x4d8ea5?, 0x4d99c4?)
	/go/src/runtime/chan.go:442 +0x12 fp=0xc000065f00 sp=0xc000065ed8 pc=0x4068f2
main.main()
	/home/ej/cgocrashbug/cgocrashbug.go:41 +0x111 fp=0xc000065f40 sp=0xc000065f00 pc=0x4ae811
runtime.main()
	/go/src/runtime/proc.go:267 +0x2bb fp=0xc000065fe0 sp=0xc000065f40 pc=0x43789b
runtime.goexit()
	/go/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc000065fe8 sp=0xc000065fe0 pc=0x464481

goroutine 2 [force gc (idle)]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
	/go/src/runtime/proc.go:398 +0xce fp=0xc000050fa8 sp=0xc000050f88 pc=0x437cee
runtime.goparkunlock(...)
	/go/src/runtime/proc.go:404
runtime.forcegchelper()
	/go/src/runtime/proc.go:322 +0xb3 fp=0xc000050fe0 sp=0xc000050fa8 pc=0x437b73
runtime.goexit()
	/go/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc000050fe8 sp=0xc000050fe0 pc=0x464481
created by runtime.init.6 in goroutine 1
	/go/src/runtime/proc.go:310 +0x1a

goroutine 3 [GC sweep wait]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
	/go/src/runtime/proc.go:398 +0xce fp=0xc000051778 sp=0xc000051758 pc=0x437cee
runtime.goparkunlock(...)
	/go/src/runtime/proc.go:404
runtime.bgsweep(0x0?)
	/go/src/runtime/mgcsweep.go:280 +0x94 fp=0xc0000517c8 sp=0xc000051778 pc=0x424014
runtime.gcenable.func1()
	/go/src/runtime/mgc.go:200 +0x25 fp=0xc0000517e0 sp=0xc0000517c8 pc=0x4193a5
runtime.goexit()
	/go/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc0000517e8 sp=0xc0000517e0 pc=0x464481
created by runtime.gcenable in goroutine 1
	/go/src/runtime/mgc.go:200 +0x66

goroutine 4 [GC scavenge wait]:
runtime.gopark(0xc00007a000?, 0x4fb148?, 0x1?, 0x0?, 0xc0000091e0?)
	/go/src/runtime/proc.go:398 +0xce fp=0xc000051f70 sp=0xc000051f50 pc=0x437cee
runtime.goparkunlock(...)
	/go/src/runtime/proc.go:404
runtime.(*scavengerState).park(0x5884a0)
	/go/src/runtime/mgcscavenge.go:425 +0x49 fp=0xc000051fa0 sp=0xc000051f70 pc=0x4218a9
runtime.bgscavenge(0x0?)
	/go/src/runtime/mgcscavenge.go:653 +0x3c fp=0xc000051fc8 sp=0xc000051fa0 pc=0x421e3c
runtime.gcenable.func2()
	/go/src/runtime/mgc.go:201 +0x25 fp=0xc000051fe0 sp=0xc000051fc8 pc=0x419345
runtime.goexit()
	/go/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc000051fe8 sp=0xc000051fe0 pc=0x464481
created by runtime.gcenable in goroutine 1
	/go/src/runtime/mgc.go:201 +0xa5

goroutine 5 [finalizer wait]:
runtime.gopark(0x4d6f40?, 0x100438e01?, 0x0?, 0x0?, 0x43fea5?)
	/go/src/runtime/proc.go:398 +0xce fp=0xc000050628 sp=0xc000050608 pc=0x437cee
runtime.runfinq()
	/go/src/runtime/mfinal.go:193 +0x107 fp=0xc0000507e0 sp=0xc000050628 pc=0x418427
runtime.goexit()
	/go/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc0000507e8 sp=0xc0000507e0 pc=0x464481
created by runtime.createfing in goroutine 1
	/go/src/runtime/mfinal.go:163 +0x3d

goroutine 6 [select]:
runtime.gopark(0xc000052690?, 0x2?, 0x8?, 0xd1?, 0xc0000525ec?)
	/go/src/runtime/proc.go:398 +0xce fp=0xc000066c38 sp=0xc000066c18 pc=0x437cee
runtime.selectgo(0xc000066e90, 0xc0000525e8, 0xc?, 0x0, 0x0?, 0x1)
	/go/src/runtime/select.go:327 +0x725 fp=0xc000066d58 sp=0xc000066c38 pc=0x447725
net.(*Resolver).lookupIPAddr(0x5881e0, {0x4fc7a8?, 0x5b5d40}, {0x4d7ae4, 0x2}, {0x4d82b6, 0x9})
	/go/src/net/lookup.go:332 +0x3fe fp=0xc000066f38 sp=0xc000066d58 pc=0x4a22fe
net.(*Resolver).LookupIPAddr(...)
	/go/src/net/lookup.go:210
net.LookupIP({0x4d82b6?, 0x0?})
	/go/src/net/lookup.go:196 +0x49 fp=0xc000066fb8 sp=0xc000066f38 pc=0x4a1d89
main.main.func1()
	/home/ej/cgocrashbug/cgocrashbug.go:26 +0x36 fp=0xc000066fe0 sp=0xc000066fb8 pc=0x4ae876
runtime.goexit()
	/go/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc000066fe8 sp=0xc000066fe0 pc=0x464481
created by main.main in goroutine 1
	/home/ej/cgocrashbug/cgocrashbug.go:23 +0xa7

goroutine 7 [select]:
runtime.gopark(0xc000067b50?, 0x2?, 0x6?, 0x0?, 0xc000067b34?)
	/go/src/runtime/proc.go:398 +0xce fp=0xc0000679e0 sp=0xc0000679c0 pc=0x437cee
runtime.selectgo(0xc000067b50, 0xc000067b30, 0x27?, 0x0, 0x436a15?, 0x1)
	/go/src/runtime/select.go:327 +0x725 fp=0xc000067b00 sp=0xc0000679e0 pc=0x447725
net.doBlockingWithCtx[...]({0x4fc818, 0xc000088050}, 0xc0000943f0)
	/go/src/net/cgo_unix.go:60 +0x14f fp=0xc000067bd0 sp=0xc000067b00 pc=0x4ad16f
net.cgoLookupIP({0x4fc818, 0xc000088050}, {0x4d7ae4, 0x2}, {0x4d82b6, 0x9})
	/go/src/net/cgo_unix.go:214 +0xb4 fp=0xc000067c00 sp=0xc000067bd0 pc=0x48f7f4
net.(*Resolver).lookupIP(0x5881e0, {0x4fc818, 0xc000088050}, {0x4d7ae4, 0x2}, {0x4d82b6, 0x9})
	/go/src/net/lookup_unix.go:70 +0x11a fp=0xc000067e58 sp=0xc000067c00 pc=0x4a373a
net.(*Resolver).lookupIP-fm({0x4fc818?, 0xc000088050?}, {0x4d7ae4?, 0x0?}, {0x4d82b6?, 0x0?})
	<autogenerated>:1 +0x49 fp=0xc000067ea0 sp=0xc000067e58 pc=0x4ae3c9
net.glob..func1({0x4fc818?, 0xc000088050?}, 0x0?, {0x4d7ae4?, 0x0?}, {0x4d82b6?, 0x0?})
	/go/src/net/hook.go:23 +0x37 fp=0xc000067ee0 sp=0xc000067ea0 pc=0x49c137
net.(*Resolver).lookupIPAddr.func1()
	/go/src/net/lookup.go:324 +0x3a fp=0xc000067f38 sp=0xc000067ee0 pc=0x4a2d1a
internal/singleflight.(*Group).doCall(0x5881f0, 0xc0000880a0, {0xc0000160c4, 0xc}, 0x0?)
	/go/src/internal/singleflight/singleflight.go:93 +0x35 fp=0xc000067fa8 sp=0xc000067f38 pc=0x48ccf5
internal/singleflight.(*Group).DoChan.func1()
	/go/src/internal/singleflight/singleflight.go:86 +0x30 fp=0xc000067fe0 sp=0xc000067fa8 pc=0x48cc90
runtime.goexit()
	/go/src/runtime/asm_amd64.s:1650 +0x1 fp=0xc000067fe8 sp=0xc000067fe0 pc=0x464481
created by internal/singleflight.(*Group).DoChan in goroutine 6
	/go/src/internal/singleflight/singleflight.go:86 +0x2e9

rax    0xb
rbx    0xe2bfc517985dee3d
rcx    0x0
rdx    0x7f9b8aa1db5c
rdi    0x7f9b8abb887b
rsi    0x7ffc4de44dc0
rbp    0x184e828
rsp    0x7ffc4de44c00
r8     0x0
r9     0x7ffc4de44c90
r10    0xc0000160d0
r11    0xb
r12    0x4f4c
r13    0x7f9b8abb887d
r14    0xb
r15    0x9
rip    0x7f9b8aa5e0cd
rflags 0x10282
cs     0x33
fs     0x0
gs     0x0
```

Built on:

```
$ cat /etc/lsb-release 
DISTRIB_ID=Ubuntu
DISTRIB_RELEASE=20.04
DISTRIB_CODENAME=focal
DISTRIB_DESCRIPTION="Ubuntu 20.04.6 LTS"
$ go version
go version go1.21.3 linux/amd64
```