//go:build windows

package main

/*
   #include <windows.h>
*/
import "C"

import "unsafe"

// Windows の伺か DLL 規約: 要求は GlobalAlloc 確保の HGLOBAL で渡され、応答も GlobalAlloc で返す。

func freeHGlobal( h unsafe.Pointer ) {
	C.GlobalFree( C.HGLOBAL( h ))
}

func allocHGlobal( size uintptr ) unsafe.Pointer {
	return unsafe.Pointer( C.GlobalAlloc( C.GPTR , C.SIZE_T( size )))
}
