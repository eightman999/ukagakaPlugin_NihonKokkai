//go:build !windows

package main

/*
   #include <stdlib.h>
*/
import "C"

import "unsafe"

// macOS など Windows 以外（Ourin）の規約: HGLOBAL の代わりに malloc 確保のポインタを使う。
// 要求はモジュールが free し、応答は malloc で確保してホストが free する。

func freeHGlobal( h unsafe.Pointer ) {
	C.free( h )
}

func allocHGlobal( size uintptr ) unsafe.Pointer {
	if size == 0 {
		size = 1
	}
	return C.malloc( C.size_t( size ))
}
