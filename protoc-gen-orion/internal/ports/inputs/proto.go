package inputs

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
