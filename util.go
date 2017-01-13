package main

import (
	"fmt"
	"runtime"
)

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func MemPrint() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	fmt.Printf("Alloc: %d\t TotalAlloc: %d\t Head: %d\t HeapSys: %d\n", mem.Alloc, mem.TotalAlloc, mem.HeapAlloc, mem.HeapSys)
}
