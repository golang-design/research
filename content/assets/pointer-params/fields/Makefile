# Copyright (c) 2020 The golang.design Initiative Authors.
# All rights reserved.
#
# The code below is produced by Changkun Ou <hi@changkun.de>.

all:
	go generate
	perflock -governor 80% go test -v -run=none -bench=. -count=10 | tee inline.txt
	benchstat inline.txt