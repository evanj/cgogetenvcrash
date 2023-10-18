CC=clang
CFLAGS:=-Wall -Wextra -Werror
EXES:=cenvleak/cenvleak

all: $(EXES)
	cd rustsetenvcrash && cargo test
	cd rustsetenvcrash && cargo clippy

cenvleak/cenvleak: cenvleak/cenvleak.c

clean:
	$(RM) $(EXES)
