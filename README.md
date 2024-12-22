# protobuf-decoder

Protobuf-decoder is POC of dynamic decoding of protocol buffer message. It accepts .proto file as an
argument together with full message name, reads protobuf message from stdin (in binary format) and
writes out this message encoded in JSON.

The dynamic decoding is useful when there's need for programmatically understand and manipulate 
Protobuf message definitions at runtime - without requiring statically generated code.

Usage:

```
cat generate-proto-message | go run main.go example/message.proto example.MyMessage
```

---

Protobuf compiler can compile the .proto file into serialized format called FileDescriptorSet. It
consists of multiple FileDescriptorProto representing the types defined in .proto file.

```
protoc --descriptor_set_out=message.pb --include_imports example/message.proto
```

## Links

* [Protobuf compiler](https://www.mankier.com/1/protoc)
* [Proto3 Language Guide](https://protobuf.dev/programming-guides/proto3/)
* [Wire format description](https://protobuf.dev/programming-guides/encoding/)
