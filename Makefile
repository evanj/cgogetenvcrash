CC=clang
CFLAGS:=-Wall -Wextra -Werror
EXES:=cenvleak/cenvleak

all: $(EXES)
	cd rustsetenvcrash && cargo test
	# https://github.com/rust-lang/rust-clippy/blob/master/README.md
	cd rustsetenvcrash && cargo clippy -- \
		--deny clippy::nursery \
		--deny clippy::pedantic

cenvleak/cenvleak: cenvleak/cenvleak.c

clean:
	$(RM) $(EXES)
