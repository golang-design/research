package main

import (
	"fmt"
	"go/types"
)

func f[T comparable]() {
	_ = map[T]T{}
}

func g[T any]() {
	// _ = map[T]T{} // ERROR: incomparable map key type T (missing comparable constraint)
}

func main() {
	_ = g[any]

	// Prior Go 1.20: any does not implement comparable
	// Since GO 1.20: OK
	_ = f[any]

	var (
		anyType = types.Universe.Lookup("any").
			Type()
		anyInterface = types.Universe.Lookup("any").
				Type().Underlying().(*types.Interface)
		comparableType = types.Universe.Lookup("comparable").
				Type()
		comparableInterface = types.Universe.Lookup("comparable").
					Type().Underlying().(*types.Interface)
	)
	fmt.Println(types.AssignableTo(comparableType, anyType))         // true
	fmt.Println(types.AssignableTo(anyType, comparableType))         // false
	fmt.Println(types.Comparable(comparableType))                    // true
	fmt.Println(types.Comparable(anyType))                           // true
	fmt.Println(types.Satisfies(comparableType, anyInterface))       // true
	fmt.Println(types.Satisfies(anyInterface, comparableInterface))  // true
	fmt.Println(types.Implements(comparableType, anyInterface))      // true
	fmt.Println(types.Implements(anyInterface, comparableInterface)) // false
}
