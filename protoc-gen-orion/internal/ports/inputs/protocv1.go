package inputs

import (
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
)

type ProtoFile struct {
	Name        string
	PackageName string
	Service     []ProtoFileService
	Location    []ProtoFileLocation
}

type ProtoFileService struct {
	Name   string
	Method []ProtoFileServiceMethod
}

type ProtoFileServiceMethod struct {
	Name            string
	ClientStreaming bool
	ServerStreaming bool
}

type ProtoFileLocation struct {
	Comments string
	Path     []int32
}

func NewProtoFileV1(file *descriptor.FileDescriptorProto) ProtoFile {
	pf := ProtoFile{
		Name:        file.GetName(),
		PackageName: strings.Replace(file.GetPackage(), ".", "_", 10),
	}
	for _, loc := range file.GetSourceCodeInfo().GetLocation() {
		if loc.LeadingComments == nil {
			continue
		}
		path := make([]int32, len(loc.GetPath()))
		copy(path, loc.GetPath())
		pfl := ProtoFileLocation{
			Comments: strings.TrimSuffix(loc.GetLeadingComments(), "\n"),
			Path:     path,
		}
		pf.Location = append(pf.Location, pfl)
	}
	for _, svc := range file.GetService() {
		pfs := ProtoFileService{
			Name: generator.CamelCase(svc.GetName()),
		}
		for _, method := range svc.GetMethod() {
			pfsm := ProtoFileServiceMethod{
				Name:            method.GetName(),
				ClientStreaming: method.GetClientStreaming(),
				ServerStreaming: method.GetServerStreaming(),
			}
			pfs.Method = append(pfs.Method, pfsm)
		}
		pf.Service = append(pf.Service, pfs)
	}
	return pf
}
