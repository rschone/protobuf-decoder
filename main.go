package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

func main() {
	if len(os.Args) < 3 {
		appName := os.Args[0]
		fmt.Fprintf(os.Stderr, "Usage: %s <proto_file> <full_message_name>\n", appName)
		fmt.Fprintf(os.Stderr, "Example: %s message.proto example.MyMessage\n", appName)
		os.Exit(1)
	}

	protoFile := os.Args[1]
	fullMessageName := os.Args[2] // example.Message

	descFilename, err := compileProto(protoFile)
	if err != nil {
		log.Fatalf("failed to compile proto: %v", err)
	}
	defer os.Remove(descFilename)

	fds, err := loadDescriptorSet(descFilename)
	if err != nil {
		log.Fatalf("failed to load descriptor set: %v", err)
	}

	msgDesc, err := findMessageDescriptor(fds, fullMessageName)
	if err != nil {
		log.Fatalf("failed to find message descriptor for %q: %v", fullMessageName, err)
	}

	// read binary protobuf message from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("failed to read stdin: %v", err)
	}

	dynamicMsg := dynamicpb.NewMessage(msgDesc)
	if err := proto.Unmarshal(data, dynamicMsg); err != nil {
		log.Fatalf("failed to unmarshal protobuf message: %v", err)
	}

	jsonData, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(dynamicMsg)
	if err != nil {
		log.Fatalf("failed to marshal to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}

func compileProto(protoFile string) (string, error) {
	descFilename := filepath.Join(os.TempDir(), "descriptor.pb")
	cmd := exec.Command("protoc", "--descriptor_set_out="+descFilename, "--include_imports", protoFile)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run protoc: %w", err)
	}
	return descFilename, nil
}

func loadDescriptorSet(descFilename string) ([]protoreflect.FileDescriptor, error) {
	data, err := os.ReadFile(descFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to read descriptor file: %w", err)
	}

	fdSet := descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(data, &fdSet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal descriptor set: %w", err)
	}

	fds := make([]protoreflect.FileDescriptor, 0, len(fdSet.File))
	for _, fd := range fdSet.File {
		fdp, err := protodesc.NewFile(fd, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create file descriptor: %w", err)
		}
		fds = append(fds, fdp)
	}

	return fds, nil
}

func findMessageDescriptor(fds []protoreflect.FileDescriptor, fullName string) (protoreflect.MessageDescriptor, error) {
	for _, fd := range fds {
		md := findMessage(fd, fullName)
		if md != nil {
			return md, nil
		}
	}
	return nil, errors.New("message not found")
}

func findMessage(fd protoreflect.FileDescriptor, fullName string) protoreflect.MessageDescriptor {
	for i := 0; i < fd.Messages().Len(); i++ {
		md := fd.Messages().Get(i)
		if string(md.FullName()) == fullName {
			return md
		}
		if found := findNestedMessage(md, fullName); found != nil {
			return found
		}
	}
	return nil
}

func findNestedMessage(md protoreflect.MessageDescriptor, fullName string) protoreflect.MessageDescriptor {
	for i := 0; i < md.Messages().Len(); i++ {
		nested := md.Messages().Get(i)
		if string(nested.FullName()) == fullName {
			return nested
		}
		if found := findNestedMessage(nested, fullName); found != nil {
			return found
		}
	}
	return nil
}
