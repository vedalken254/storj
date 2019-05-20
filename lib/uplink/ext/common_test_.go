// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

// #cgo CFLAGS: -g -Wall
// #include <stdbool.h>
// #include "c/tests/test.h"
// #include "c/headers/main.h"
import "C"
import (
	"encoding/json"
	"github.com/nsf/jsondiff"
	"unsafe"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"storj.io/storj/lib/uplink/ext/pb"
	"storj.io/storj/lib/uplink/ext/testing"
)

func TestGoToCStruct_success(t *testing.T) {
	{
		t.Info("go to C string")

		stringGo := "testing 123"
		toCString := C.CString("")

		err := GoToCStruct(stringGo, &toCString)
		require.NoError(t, err)

		assert.Equal(t, stringGo, C.GoString(toCString))
	}

	{
		t.Info("go to C bool")

		boolGo := true
		var toCBool C.bool

		err := GoToCStruct(boolGo, &toCBool)
		require.NoError(t, err)

		assert.Equal(t, boolGo, bool(toCBool))
	}

	{
		t.Info("go to C simple struct")

		simpleGo := simple{"one", -2, 3,}
		toCStruct := C.struct_Simple{}

		err := GoToCStruct(simpleGo, &toCStruct)
		require.NoError(t, err)

		assert.Equal(t, simpleGo.Str1, C.GoString(toCStruct.Str1))
		assert.Equal(t, simpleGo.Int2, int(toCStruct.Int2))
		assert.Equal(t, simpleGo.Uint3, uint(toCStruct.Uint3))
	}

	{
		t.Info("go to C nested struct")

		simpleGo := simple{"two", -10, 5,}
		nestedGo := nested{simpleGo, 4}
		toCStruct := C.struct_Nested{}

		err := GoToCStruct(nestedGo, &toCStruct)
		require.NoError(t, err)

		assert.Equal(t, nestedGo.Simple.Str1, C.GoString(toCStruct.Simple.Str1))
		assert.Equal(t, nestedGo.Simple.Int2, int(toCStruct.Simple.Int2))
		assert.Equal(t, nestedGo.Simple.Uint3, uint(toCStruct.Simple.Uint3))
		assert.Equal(t, nestedGo.Int4, int(toCStruct.Int4))
	}
}

func TestGoToCStruct_error(t *testing.T) {
	// TODO
}

func TestCToGoStruct_success(t *testing.T) {
	{
		t.Info("C to go string")

		stringC := C.CString("testing 123")
		toGoString := ""

		err := CToGoStruct(stringC, &toGoString)
		require.NoError(t, err)

		assert.Equal(t, C.GoString(stringC), toGoString)
	}

	{
		t.Info("C to go bool")

		boolC := C.bool(true)
		toGoBool := false

		err := CToGoStruct(boolC, &toGoBool)
		require.NoError(t, err)

		assert.Equal(t, bool(boolC), toGoBool)
	}

	{
		t.Info("C to go simple struct")

		simpleC := C.struct_Simple{C.CString("one"), -2, 3,}
		toGoStruct := simple{}

		err := CToGoStruct(simpleC, &toGoStruct)
		require.NoError(t, err)

		assert.Equal(t, C.GoString(simpleC.Str1), toGoStruct.Str1)
		assert.Equal(t, int(simpleC.Int2), toGoStruct.Int2)
		assert.Equal(t, uint(simpleC.Uint3), toGoStruct.Uint3)
	}
}

func TestCToGoStruct_error(t *testing.T) {
	// TODO
}

func TestSendToGo_success(t *testing.T) {
	{
		t.Info("uplink config")

		startConfig := &pb.UplinkConfig{
			// -- WIP | TODO
			Tls: &pb.TLSConfig{
				SkipPeerCaWhitelist: true,
				PeerCaWhitelistPath: "/whitelist.pem",
			},
			IdentityVersion: &pb.IDVersion{
				Number: 0,
			},
			MaxInlineSize: 1,
			MaxMemory:     2,
		}
		snapshot, err := proto.Marshal(startConfig)
		require.NoError(t, err)
		require.NotEmpty(t, snapshot)

		// NB/TODO: I don't think this is exactly right but might work
		size := uintptr(len(snapshot))

		t.Info("", zap.ByteString("snapshot 1", snapshot))
		cVal := &C.struct_GoValue{
			//Ptr: (0 by default),
			Type:     C.UplinkConfigType,
			Snapshot: (*C.uchar)(unsafe.Pointer(&snapshot)),
			Size:     C.ulong(size),
		}
		t.Info("", zap.ByteString("snapshot 2", *(*[]byte)(unsafe.Pointer(cVal.Snapshot))))
		assert.Zero(t, cVal.Ptr)

		cErr := C.CString("")
		SendToGo(cVal, &cErr)
		//t.Info("", zap.ByteString("snapshot 3", *(*[]byte)(unsafe.Pointer(cVal.Snapshot))))
		require.Empty(t, C.GoString(cErr))

		assert.NotZero(t, uintptr(cVal.Ptr))
		assert.NotZero(t, cVal.Type)

		//value := CToGoGoValue(*cVal)
		//endConfig := structRefMap.Get(token(value.ptr))
		//require.Equal(t, uintptr(cVal.Ptr), value.ptr)
		endConfig := structRefMap.Get(token(cVal.Ptr))

		startJSON, err := json.Marshal(startConfig)
		require.NoError(t, err)

		endJSON, err := json.Marshal(endConfig)
		require.NoError(t, err)

		match, diffStr := jsondiff.Compare(startJSON, endJSON, &jsondiff.Options{})
		if !assert.Equal(t, jsondiff.FullMatch, match) {
			t.Error("config JSON diff:", zap.String("", diffStr))
		}
	}

	// TODO: other types
}

func TestSendToGo_error(t *testing.T) {
	// TODO
}

func TestCToGoGoValue(t *testing.T) {
	testMap := newMapping()
	str := "test string 123"
	strToken := testMap.Add(str)
	cVal := C.struct_GoValue{
		Ptr: C.GoUintptr(strToken),
		// NB: arbitrary type
		Type: C.APIKeyType,
	}

	value := CToGoGoValue(cVal)
	assert.Equal(t, uint(cVal.Type), value._type)
	assert.NotZero(t, value.ptr)

	gotStr := testMap.Get(token(value.ptr))
	assert.Equal(t, str, gotStr)
}

func TestMapping_Add(t *testing.T) {
	{
		t.Info("string")
		testMap := newMapping()

		str := "testing 123"
		strToken := testMap.Add(str)

		gotStr, ok := testMap.values[strToken]
		require.True(t, ok)
		assert.Equal(t, str, gotStr)
	}

	{
		t.Info("pointer")
		testMap := newMapping()

		str := "testing 123"
		strToken := testMap.Add(&str)

		gotStr, ok := testMap.values[strToken]
		require.True(t, ok)
		assert.Equal(t, str, *gotStr.(*string))
	}
}

func TestMapping_Get(t *testing.T) {
	{
		t.Info("string")
		testMap := newMapping()

		str := "testing 123"
		strToken := token(1)
		testMap.values[strToken] = str

		gotStr := testMap.Get(strToken)
		assert.Equal(t, str, gotStr)
	}

	{
		t.Info("pointer")
		testMap := newMapping()

		str := "testing 123"
		strToken := token(1)
		testMap.values[strToken] = &str

		gotStr := testMap.Get(strToken)
		assert.Equal(t, str, *gotStr.(*string))
	}
}
