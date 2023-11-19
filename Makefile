CC=clang
CFLAGS:=-Wall -Wextra -Werror
EXES:=cenvleak/cenvleak environvar/environvar

all: $(EXES)
	cd rustsetenvcrash && cargo test
	# https://github.com/rust-lang/rust-clippy/blob/master/README.md
	cd rustsetenvcrash && cargo clippy -- \
		--deny clippy::nursery \
		--deny clippy::pedantic

	# DeprecatedOrUnsafeBufferHandling: Warns about snprintf which has no alternative in glibc
	clang-tidy \
		--checks=all,-clang-analyzer-security.insecureAPI.DeprecatedOrUnsafeBufferHandling \
		cenvleak/cenvleak.c
	
	go test ./...
	go vet ./...
	staticcheck --checks=all ./...
	goimports -w .

cenvleak/cenvleak: cenvleak/cenvleak.c

environvar/environvar: environvar/environvar.c

clean:
	$(RM) $(EXES)
