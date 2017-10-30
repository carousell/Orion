// protoc-gen-go is a plugin for the Google protocol buffer compiler to generate
// Go code.  Run it by building this program and putting it in your path with
// the name
// 	protoc-gen-go
// That word 'go' at the end becomes part of the option string set for the
// protocol compiler, so once the protocol compiler (protoc) is installed
// you can run
// 	protoc --go_out=output_directory input_directory/file.proto
// to generate Go bindings for the protocol defined by file.proto.
// With that input, the output will be written to
// 	output_directory/file.pb.go
//
// The generated code is documented in the package comment for
// the library.
//
// See the README and documentation for protocol buffers to learn more:
// 	https://developers.google.com/protocol-buffers/
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

var (
	grpcPkg = generator.RegisterUniquePackageName("grpc", nil)
)

// P prints the arguments to the generated output.  It handles strings and int32s, plus
// handling indirections because they may be *string, etc.
func P(g *generator.Generator, str ...interface{}) {
	//g.WriteString(g.indent)
	for _, v := range str {
		switch s := v.(type) {
		case string:
			g.WriteString(s)
		case *string:
			g.WriteString(*s)
		case bool:
			fmt.Fprintf(g, "%t", s)
		case *bool:
			fmt.Fprintf(g, "%t", *s)
		case int:
			fmt.Fprintf(g, "%d", s)
		case *int32:
			fmt.Fprintf(g, "%d", *s)
		case *int64:
			fmt.Fprintf(g, "%d", *s)
		case float64:
			fmt.Fprintf(g, "%g", s)
		case *float64:
			fmt.Fprintf(g, "%g", *s)
		default:
			g.Fail(fmt.Sprintf("unknown type in printer: %T", v))
		}
	}
	g.WriteByte('\n')
}

// Generate the package definition
func generateHeader(g *generator.Generator, file *descriptor.FileDescriptorProto) {
	P(g, "// Code generated by protoc-gen-orion. DO NOT EDIT.")
	P(g, "// source: ", file.Name)
	P(g)

	name := file.GetPackage()

	P(g, "package ", name)
	P(g, "")
	P(g, "import (")
	P(g, "\torion \"github.com/carousell/Orion/orion\"")
	P(g, ")")
	P(g, "")
}

// Generate the file
func generate(g *generator.Generator, file *descriptor.FileDescriptorProto) {
	generateHeader(g, file)
	for _, svc := range file.GetService() {
		origServName := svc.GetName()
		fullServName := origServName
		if pkg := file.GetPackage(); pkg != "" {
			fullServName = pkg + "." + fullServName
		}
		servName := generator.CamelCase(origServName)
		serviceDescVar := "_" + servName + "_serviceDesc"
		serverType := servName + "Server"

		P(g, "func Register", servName, "OrionServer(srv ", serverType, ", orionServer orion.Server) {")
		//g.P("s.RegisterService(&", serviceDescVar, `, srv)`)
		P(g, "\torionServer.RegisterService(&", serviceDescVar, `, srv)`)
		P(g, "}")
		P(g)
	}
}

func main() {
	g := generator.New()

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		g.Error(err, "reading input")
	}

	if err := proto.Unmarshal(data, g.Request); err != nil {
		g.Error(err, "parsing input proto")
	}

	if len(g.Request.FileToGenerate) == 0 {
		g.Fail("no files to generate")
	}

	for _, file := range g.Request.GetProtoFile() {
		g.Reset()
		generate(g, file)
		g.Response.File = append(g.Response.File, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(strings.ToLower(file.GetName()) + ".orion.pb.go"),
			Content: proto.String(g.String()),
		})
	}

	// Send back the results.
	data, err = proto.Marshal(g.Response)
	if err != nil {
		g.Error(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		g.Error(err, "failed to write output proto")
	}
}
