package main

/*
   #include <stdlib.h>
   #include <string.h>
*/
import "C"

import (
    "fmt"
    "unsafe"
    "strings"
    "regexp"
)



func main() {
    fmt.Println( "test" )
}

var Sec         int
var Interval    int

var Directory    string
var References []string 
var CheckID         = regexp.MustCompile("^ID: ")
var CheckReference  = regexp.MustCompile("^Reference.+?: ")

type ResponseStruct struct {
    Shiori  string
    Sender  string
    Charset string
    Marker  string
    Value   string
}
func GetResponse( r *ResponseStruct ) string {
    V := ""
    if r.Value  != "" { V = "Value: "  + r.Value     + "\r\n" }
    res :=  r.Shiori    + "\r\n" + 
            r.Sender    + "\r\n" + 
            r.Charset   + "\r\n" + 
            V + "\r\n\r\n"
    return res
}


//export load
func load(h unsafe.Pointer, length C.long ) C.int {
    fmt.Println( "load NihonKokkai" )
    Directory = C.GoStringN(( *C.char )( unsafe.Pointer( h )), ( C.int )( length ))
    fmt.Println( Directory  )

    //設定読み込み
    LoadJson()

    ////初期化。
    Sec         = 0
    Interval    = Config.IntervalSec
    OldKokkai   = ""

    ////国会回収
    go do()

	freeHGlobal( h )
	return 1
}


//export unload
func unload() C.int {
    fmt.Println( "unload NihonKokkai" )
	return 1
}


//export request
func request( h unsafe.Pointer, length *C.long ) unsafe.Pointer {
	RequestText := C.GoStringN(( *C.char )( unsafe.Pointer( h )), ( C.int )( *length ))
	freeHGlobal( h )


    Value           := ""
    ID              := ""
    References      = []string{}
    //var NOTIFY bool = false

    Response := new( ResponseStruct )
    Response.Sender  = "Sender: NihonKokkai"
    Response.Charset = "Charset: UTF-8"

    //IDとReference
    //必要な情報を分解する。
    RequestLines := strings.Split( RequestText , "\r\n" )
    for _ , line := range RequestLines {
        if( line == "NOTIFY PLUGIN/2.0" ){
            //"GET PLUGIN/2.0";
            //NOTIFY = true

        } else if CheckID.MatchString( line )  {
            //fmt.Println( line )
            ID = CheckID.ReplaceAllString( line , "" )

        } else if CheckReference.MatchString( line )  {
            //fmt.Println( line )
            ref := CheckReference.ReplaceAllString( line , "" )
            References = append( References , ref )

        } else {
            //fmt.Println( line )
        }
    }

    //実行関数
    if ID == "OnOtherGhostTalk" {
    //} else if ID == "OnSecondChange" {
    } else if ID == "OnSecondChange" && len( KokkaiArray ) != 0 {
        //実行開始
        if Config.StartSec < Sec {
            if Interval >= Config.IntervalSec {
                Interval = 0
                Value           = "\\0\\b[2]\\_q" + KokkaiArray[0] + "\\_q"
                OldKokkai       = "\\0\\b[2]\\_q" + KokkaiArray[0] + "\\_q"
                KokkaiArray     = KokkaiArray[1:]
            }
            Interval++
        } else {
            Sec++
        }

    } else if ID == "OnMenuExec"  {
    } else if ID == "OnKokkaiUrl"  {
        Value = "\\j[" + References[0] + "]" + OldKokkai
    } else {
        //fmt.Println( "no touch :" + ID )
        //fmt.Print( "NOTIFY : " )
        //fmt.Println( NOTIFY )
        //fmt.Print( "References : " )
        //fmt.Println( References )
        //fmt.Println( "" )
    }


    if Value == "" {
        Response.Shiori  = "PLUGIN/2.0 204 No Content"
    } else {
        Response.Shiori = "PLUGIN/2.0 200 OK"
        Response.Value  = Value
    }

    res_buf := C.CString( GetResponse( Response ))
    defer C.free( unsafe.Pointer( res_buf ))

	res_size := C.strlen( res_buf )
	ret      := allocHGlobal( uintptr( res_size ))
	C.memcpy(( unsafe.Pointer )( ret ) , ( unsafe.Pointer )( res_buf ) , res_size )
	*length = ( C.long )( res_size )
	return ret
}


















