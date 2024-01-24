package main

import (
	"fmt"
	"net"

	"github.com/jxskiss/gopkg/v2/infra/logid"
)

func main() {
	N := 3

	// default v1 generator
	fmt.Println("v1 default generator:")
	{
		id1 := logid.Gen()
		fmt.Println(id1)
		fmt.Println(logid.Decode(id1))
		for i := 0; i < N; i++ {
			fmt.Println(logid.Gen())
		}
	}

	// v2 generator
	for _, ip := range []string{
		"1.2.3.4",
		"fdbd:dc01:16:16::94",
		"",
	} {
		fmt.Printf("\nv2 generator, ip: %q\n", ip)
		v2Gen := logid.NewV2Gen(net.ParseIP(ip))
		id2 := v2Gen.Gen()
		fmt.Println(id2)
		fmt.Println(logid.Decode(id2))
		for i := 0; i < N; i++ {
			fmt.Println(v2Gen.Gen())
		}
	}
}
