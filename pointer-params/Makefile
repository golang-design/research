GOVERSION=$(shell go version | awk '{print $$3}')
all:
	perflock -governor 80% go test -v -run=none -bench=. -count=10 | tee $(GOVERSION).txt
	benchstat $(GOVERSION).txt
asm:
	mkdir asm && go tool compile -S vec.go > asm/vec.s