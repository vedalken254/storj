/* Generated by the protocol buffer compiler.  DO NOT EDIT! */
/* Generated from: uplink.proto */

/* Do not generate deprecated warnings for self */
#ifndef PROTOBUF_C__NO_DEPRECATED
#define PROTOBUF_C__NO_DEPRECATED
#endif

#include "uplink.pb-c.h"
void   storj__libuplink__idversion__init
                     (Storj__Libuplink__IDVersion         *message)
{
  static const Storj__Libuplink__IDVersion init_value = STORJ__LIBUPLINK__IDVERSION__INIT;
  *message = init_value;
}
size_t storj__libuplink__idversion__get_packed_size
                     (const Storj__Libuplink__IDVersion *message)
{
  assert(message->base.descriptor == &storj__libuplink__idversion__descriptor);
  return protobuf_c_message_get_packed_size ((const ProtobufCMessage*)(message));
}
size_t storj__libuplink__idversion__pack
                     (const Storj__Libuplink__IDVersion *message,
                      uint8_t       *out)
{
  assert(message->base.descriptor == &storj__libuplink__idversion__descriptor);
  return protobuf_c_message_pack ((const ProtobufCMessage*)message, out);
}
size_t storj__libuplink__idversion__pack_to_buffer
                     (const Storj__Libuplink__IDVersion *message,
                      ProtobufCBuffer *buffer)
{
  assert(message->base.descriptor == &storj__libuplink__idversion__descriptor);
  return protobuf_c_message_pack_to_buffer ((const ProtobufCMessage*)message, buffer);
}
Storj__Libuplink__IDVersion *
       storj__libuplink__idversion__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data)
{
  return (Storj__Libuplink__IDVersion *)
     protobuf_c_message_unpack (&storj__libuplink__idversion__descriptor,
                                allocator, len, data);
}
void   storj__libuplink__idversion__free_unpacked
                     (Storj__Libuplink__IDVersion *message,
                      ProtobufCAllocator *allocator)
{
  if(!message)
    return;
  assert(message->base.descriptor == &storj__libuplink__idversion__descriptor);
  protobuf_c_message_free_unpacked ((ProtobufCMessage*)message, allocator);
}
static const ProtobufCFieldDescriptor storj__libuplink__idversion__field_descriptors[2] =
{
  {
    "number",
    1,
    PROTOBUF_C_LABEL_NONE,
    PROTOBUF_C_TYPE_UINT32,
    0,   /* quantifier_offset */
    offsetof(Storj__Libuplink__IDVersion, number),
    NULL,
    NULL,
    0,             /* flags */
    0,NULL,NULL    /* reserved1,reserved2, etc */
  },
  {
    "new_private_key",
    2,
    PROTOBUF_C_LABEL_NONE,
    PROTOBUF_C_TYPE_UINT64,
    0,   /* quantifier_offset */
    offsetof(Storj__Libuplink__IDVersion, new_private_key),
    NULL,
    NULL,
    0,             /* flags */
    0,NULL,NULL    /* reserved1,reserved2, etc */
  },
};
static const unsigned storj__libuplink__idversion__field_indices_by_name[] = {
  1,   /* field[1] = new_private_key */
  0,   /* field[0] = number */
};
static const ProtobufCIntRange storj__libuplink__idversion__number_ranges[1 + 1] =
{
  { 1, 0 },
  { 0, 2 }
};
const ProtobufCMessageDescriptor storj__libuplink__idversion__descriptor =
{
  PROTOBUF_C__MESSAGE_DESCRIPTOR_MAGIC,
  "storj.libuplink.IDVersion",
  "IDVersion",
  "Storj__Libuplink__IDVersion",
  "storj.libuplink",
  sizeof(Storj__Libuplink__IDVersion),
  2,
  storj__libuplink__idversion__field_descriptors,
  storj__libuplink__idversion__field_indices_by_name,
  1,  storj__libuplink__idversion__number_ranges,
  (ProtobufCMessageInit) storj__libuplink__idversion__init,
  NULL,NULL,NULL    /* reserved[123] */
};
