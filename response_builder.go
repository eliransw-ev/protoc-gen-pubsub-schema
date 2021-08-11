package main

import (
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type responseBuilder struct {
	request      *pluginpb.CodeGeneratorRequest
	protoFiles   map[string]*descriptorpb.FileDescriptorProto
	messageTypes map[string]*descriptorpb.DescriptorProto
}

func buildResponseError(errorMessage string) *pluginpb.CodeGeneratorResponse {
	return &pluginpb.CodeGeneratorResponse{Error: &errorMessage}
}

func newResponseBuilder(req *pluginpb.CodeGeneratorRequest) responseBuilder {
	builder := responseBuilder{
		req,
		make(map[string]*descriptorpb.FileDescriptorProto),
		make(map[string]*descriptorpb.DescriptorProto),
	}
	builder.initIndex()
	return builder
}

func (b *responseBuilder) initIndex() {
	for _, protoFile := range b.request.GetProtoFile() {
		b.protoFiles[protoFile.GetName()] = protoFile
		packageName := strings.TrimSuffix("."+protoFile.GetPackage(), ".")
		b.initIndexByMessages(packageName, protoFile.GetMessageType())
	}
}

func (b *responseBuilder) initIndexByMessages(messageNamePrefix string, messages []*descriptorpb.DescriptorProto) {
	for _, message := range messages {
		messageName := messageNamePrefix + "." + message.GetName()
		b.messageTypes[messageName] = message
		b.initIndexByMessages(messageName, message.GetNestedType())
	}
}

func (b responseBuilder) build() *pluginpb.CodeGeneratorResponse {
	resp := new(pluginpb.CodeGeneratorResponse)
	for _, fileName := range b.request.GetFileToGenerate() {
		respFile, err := b.buildFile(fileName)
		if err != nil {
			return buildResponseError(err.Error())
		}
		resp.File = append(resp.File, respFile)
	}
	return resp
}

func (b responseBuilder) buildFile(reqFileName string) (*pluginpb.CodeGeneratorResponse_File, error) {
	respFileName := strings.TrimSuffix(reqFileName, ".proto") + ".pubsub.proto"
	content, err := newContentBuilder(b).build(b.protoFiles[reqFileName])
	return &pluginpb.CodeGeneratorResponse_File{
		Name:    &respFileName,
		Content: &content,
	}, err
}
