all:
	go generate
	perflock -governor 80% go test -v -run=none -bench=. -count=10 | tee inline.txt
	benchstat inline.txt