package main

import "math"

func main() {
	m := map[any]any{}
	println(len(m)) // 0
	m[math.NaN()] = struct{}{}
	println(len(m)) // 1
	for x := range m {
		delete(m, x)
	}
	println(len(m)) // 1
	delete(m, math.NaN())
	println(len(m))
}
