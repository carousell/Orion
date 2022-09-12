package inputs

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

type ProtoRequest struct {
	ProtoFile   ProtoFile
	ProtoParams ProtoParams
}

type ProtoFile struct {
	Name          string
	PackageName   string
	GoPackagePath string
	Service       []ProtoFileService
	Location      []ProtoFileLocation
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

type ProtoParams struct {
	StandAloneMode      bool
	ExportedServiceDesc bool
}

func NewProtoFile(file *descriptorpb.FileDescriptorProto) ProtoFile {
	pf := ProtoFile{
		Name:          file.GetName(),
		PackageName:   strings.Replace(file.GetPackage(), ".", "_", 10),
		GoPackagePath: parseGoPackage(file.GetOptions().GetGoPackage()),
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

func NewProtoParams(reqParam string) (ProtoParams, error) {
	pp := ProtoParams{}
	for _, param := range strings.Split(reqParam, ",") {
		var value string
		if i := strings.Index(param, "="); i >= 0 {
			value = param[i+1:]
			param = param[0:i]
		}
		switch param {
		case "standalone-mode":
			boolValue, err := parseBoolValue(value)
			if err != nil {
				return ProtoParams{}, fmt.Errorf(`bad value for parameter %q: %w`, param, err)
			}
			pp.StandAloneMode = boolValue
		case "exported-service-desc":
			boolValue, err := parseBoolValue(value)
			if err != nil {
				return ProtoParams{}, fmt.Errorf(`bad value for parameter %q: %w`, param, err)
			}
			pp.ExportedServiceDesc = boolValue
		}
	}
	return pp, nil
}

func parseGoPackage(goPackage string) string {
	parts := strings.Split(goPackage, ";")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func parseBoolValue(value string) (bool, error) {
	switch value {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf(`want "true" or "false"`)
	}
}
