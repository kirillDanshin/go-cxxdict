package main

import (
	"fmt"

	// 3rd-party
	"mylib"
)

func main() {

	fmt.Printf("mylib.NewD1(\"d1\")...\n")
	d1 := mylib.NewD1("d1")
	fmt.Printf("mylib.NewD1(\"d1\")...[ok]\n")

	n := d1.Name()
	fmt.Printf("d1.Name() = \"%s\"\n", n)

	fmt.Printf("d1.Do_hello(\"you\")...\n")
	d1.Do_hello("you")

	fmt.Printf("d1.Do_virtual_hello(\"you\")...\n")
	d1.Do_virtual_hello("you")

	fmt.Printf("d1.Pure_virtual_method(\"you\")...\n")
	d1.Pure_virtual_method("you")

	var b mylib.Base

	fmt.Printf("\n/// test implicit conversion to base-class' interface\n")
	fmt.Printf("call d1 methods via mylib.Base...\n")
	b = d1

	fmt.Printf("b.Do_hello(\"you\")...\n")
	b.Do_hello("you")

	fmt.Printf("b.Do_virtual_hello(\"you\")...\n")
	b.Do_virtual_hello("you")

	fmt.Printf("b.Pure_virtual_method(\"you\")...\n")
	b.Pure_virtual_method("you")

	fmt.Printf("call d1 methods via mylib.Base...[done]\n")

	fmt.Printf("\n/// now, re-test but using an explicit call to GocxxGet<base-class>()\n")
	fmt.Printf("call d1 methods via mylib.Base...\n")
	b = d1.GocxxGetBase()

	fmt.Printf("b.Do_hello(\"you\")...\n")
	b.Do_hello("you")

	fmt.Printf("b.Do_virtual_hello(\"you\")...\n")
	b.Do_virtual_hello("you")

	fmt.Printf("b.Pure_virtual_method(\"you\")...\n")
	b.Pure_virtual_method("you")

	fmt.Printf("call d1 methods via mylib.Base...[done]\n")

	mylib.DeleteD1(d1)

	fmt.Printf("mylib.NewD1(\"d12\")...\n")
	d12 := mylib.NewD1("d12")
	fmt.Printf("mylib.NewD1(\"d12\")...[ok]\n")

	fmt.Printf("delete d12 via ~Base...\n")
	mylib.DeleteBase(d12)
	fmt.Printf("delete d12 via ~Base...[ok]\n")

}
