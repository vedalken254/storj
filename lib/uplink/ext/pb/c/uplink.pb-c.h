/* Generated by the protocol buffer compiler.  DO NOT EDIT! */
/* Generated from: uplink.proto */

#ifndef PROTOBUF_C_uplink_2eproto__INCLUDED
#define PROTOBUF_C_uplink_2eproto__INCLUDED

#include <protobuf-c/protobuf-c.h>

PROTOBUF_C__BEGIN_DECLS

#if PROTOBUF_C_VERSION_NUMBER < 1003000
# error This file was generated by a newer version of protoc-c which is incompatible with your libprotobuf-c headers. Please update your headers.
#elif 1003001 < PROTOBUF_C_MIN_COMPILER_VERSION
# error This file was generated by an older version of protoc-c which is incompatible with your libprotobuf-c headers. Please regenerate this file with a newer version of protoc-c.
#endif


typedef struct _Storj__Libuplink__IDVersion Storj__Libuplink__IDVersion;


/* --- enums --- */


/* --- messages --- */

struct  _Storj__Libuplink__IDVersion
{
  ProtobufCMessage base;
  uint32_t number;
  uint64_t new_private_key;
};
#define STORJ__LIBUPLINK__IDVERSION__INIT \
 { PROTOBUF_C_MESSAGE_INIT (&storj__libuplink__idversion__descriptor) \
    , 0, 0 }


/* Storj__Libuplink__IDVersion methods */
void   storj__libuplink__idversion__init
                     (Storj__Libuplink__IDVersion         *message);
size_t storj__libuplink__idversion__get_packed_size
                     (const Storj__Libuplink__IDVersion   *message);
size_t storj__libuplink__idversion__pack
                     (const Storj__Libuplink__IDVersion   *message,
                      uint8_t             *out);
size_t storj__libuplink__idversion__pack_to_buffer
                     (const Storj__Libuplink__IDVersion   *message,
                      ProtobufCBuffer     *buffer);
Storj__Libuplink__IDVersion *
       storj__libuplink__idversion__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data);
void   storj__libuplink__idversion__free_unpacked
                     (Storj__Libuplink__IDVersion *message,
                      ProtobufCAllocator *allocator);
/* --- per-message closures --- */

typedef void (*Storj__Libuplink__IDVersion_Closure)
                 (const Storj__Libuplink__IDVersion *message,
                  void *closure_data);

/* --- services --- */


/* --- descriptors --- */

extern const ProtobufCMessageDescriptor storj__libuplink__idversion__descriptor;

PROTOBUF_C__END_DECLS


#endif  /* PROTOBUF_C_uplink_2eproto__INCLUDED */
