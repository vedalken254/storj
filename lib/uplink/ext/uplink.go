// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

// #cgo CFLAGS: -g -Wall
// #ifndef UPLINK_HEADERS
//   #define UPLINK_HEADERS
//   #include "headers/main.h"
// #endif
import "C"
import (
	"context"
	"fmt"
	"unsafe"

	monkit "gopkg.in/spacemonkeygo/monkit.v2"

	"storj.io/storj/lib/uplink"
)

var mon = monkit.Package()

//export NewUplink
func NewUplink(ctxPtr C.GoCtxPtr, cConfig C.struct_Config, cErr **C.char) (cUplink C.struct_Uplink) {
	goConfig := new(uplink.Config)
	if err := CToGoStruct(cConfig, goConfig); err != nil {
		*cErr = C.CString(err.Error())
		return cUplink
	}

	ctx, ok := ctxMapping.Get(token(ctxPtr)).(context.Context)
	if !ok {
		*cErr = C.CString("Could not retrieve context.Context from context Map")
		return cUplink
	}

	goUplink, err := uplink.NewUplink(ctx, goConfig)
	if err != nil {
		fmt.Printf("NewUplink go err: %s\n", err)
		*cErr = C.CString(err.Error())
		return cUplink
	}

	return C.struct_Uplink{
		GoUplink: cPointerFromGoStruct(goUplink),
		Config:   cConfig,
	}
}

//export OpenProject
func OpenProject(cUplink C.struct_Uplink, satelliteAddr *C.char, cAPIKey C.APIKey, cOpts C.struct_ProjectOptions, cErr **C.char) (cProject C.Project) {
	var err error
	ctx := context.Background()
	defer mon.Task()(&ctx)(&err)

	goUplink := (*uplink.Uplink)(unsafe.Pointer(uintptr(cUplink.GoUplink)))

	opts := new(uplink.ProjectOptions)
	err = CToGoStruct(cOpts, opts)
	if err != nil {
		*cErr = C.CString(err.Error())
		fmt.Println(cErr, err.Error())
		return cProject
	}

	apiKey := (*uplink.APIKey)(goPointerFromCGoUintptr(cAPIKey))

	project, err := goUplink.OpenProject(ctx, C.GoString(satelliteAddr), *apiKey, opts)
	if err != nil {
		*cErr = C.CString(err.Error())
		fmt.Println(cErr, err.Error())
		return cProject
	}
	return cPointerFromGoStruct(project)
}