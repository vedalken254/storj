// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

// #cgo CFLAGS: -g -Wall
// #ifndef STORJ_HEADERS
//   #define STORJ_HEADERS
//   #include "c/headers/main.h"
// #endif
import "C"
import (
	"context"

	"gopkg.in/spacemonkeygo/monkit.v2"

	"storj.io/storj/lib/uplink"
)

var mon = monkit.Package()

//export NewUplink
func NewUplink(configRef C.UplinkConfigRef, cErr **C.char) (cUplink C.gvUplink) {
	config, ok := structRefMap.Get(token(configRef)).(*uplink.Config)
	if !ok {
		*cErr = C.CString("invalid config")
		return cUplink
	}

	goUplink, err := uplink.NewUplink(context.Background(), config)
	if err != nil {
		*cErr = C.CString(err.Error())
		return cUplink
	}

	return C.gvUplink{
		Ptr: C.UplinkRef(structRefMap.Add(goUplink)),
		Type: C.UplinkType,
	}
}

//export OpenProject
func OpenProject(cUplink C.UplinkRef, satelliteAddr *C.char, cAPIKey C.APIKeyRef, cOpts C.struct_ProjectOptions, cErr **C.char) (cProject C.Project) {
	var err error
	ctx := context.Background()
	defer mon.Task()(&ctx)(&err)

	goUplink, ok := structRefMap.Get(token(cUplink)).(*uplink.Uplink)
	if !ok {
		*cErr = C.CString("invalid uplink")
		return cProject	}

	opts := new(uplink.ProjectOptions)
	err = CToGoStruct(cOpts, opts)
	if err != nil {
		*cErr = C.CString(err.Error())
		fmt.Println(cErr, err.Error())
		return cProject
	}

	apiKey, ok := structRefMap.Get(token(cAPIKey)).(uplink.APIKey)
	if !ok {
		*cErr = C.CString("invalid API Key")
		fmt.Println(cErr, err.Error())
		return cProject
	}

	project, err := goUplink.OpenProject(ctx, C.GoString(satelliteAddr), apiKey, opts)
	if err != nil {
		*cErr = C.CString(err.Error())
		fmt.Println(cErr, err.Error())
		return cProject
	}
	return cPointerFromGoStruct(project)
	return cProject
}
