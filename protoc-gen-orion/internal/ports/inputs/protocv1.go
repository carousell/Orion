package inputs

import (
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

func NewProtoFileV1(file *descriptorpb.FileDescriptorProto) ProtoFile {
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
			Name: svc.GetName(),
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
