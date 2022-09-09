// protoc-gen-orion is a plugin for the Google protocol buffer compiler to generate
// Orion Go code.  Run it by building this program and putting it in your path with
// the name
// 	protoc-gen-orion
//
// The generated code is documented in the package comment for
// the library.
package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/carousell/Orion/protoc-gen-orion/internal/generator"
	"github.com/carousell/Orion/protoc-gen-orion/internal/ports/inputs"
	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// Error reports a problem, including an error, and exits the program.
func logError(err error, msgs ...string) {
	s := strings.Join(msgs, " ") + ":" + err.Error()
	log.Print("protoc-gen-orion: error:", s)
	os.Exit(1)
}

// Fail reports a problem and exits the program.
func logFail(msgs ...string) {
	s := strings.Join(msgs, " ")
	log.Print("protoc-gen-orion: error:", s)
	os.Exit(1)
}

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		logError(err, "reading input")
	}

	request := new(plugin.CodeGeneratorRequest)
	if err := proto.Unmarshal(data, request); err != nil {
		logError(err, "parsing input proto")
	}

	if len(request.FileToGenerate) == 0 {
		logFail("no files to generate")
	}
	filesToGenerate := make(map[string]bool)
	for _, v := range request.FileToGenerate {
		filesToGenerate[v] = true
	}

	response := new(plugin.CodeGeneratorResponse)
	response.File = make([]*plugin.CodeGeneratorResponse_File, 0)

	for _, file := range request.GetProtoFile() {
		if _, ok := filesToGenerate[file.GetName()]; ok {
			// check if file has any service
			if len(file.Service) > 0 {
				gr, err := generator.GenerateFile(inputs.NewProtoFileV1(file))
				if err != nil {
					logError(err, "failed to generate file")
					continue
				}
				f := &plugin.CodeGeneratorResponse_File{
					Name:    proto.String(gr.Name),
					Content: proto.String(gr.Content),
				}
				response.File = append(response.File, f)
			}
		}
	}

	// Send back the results.
	data, err = proto.Marshal(response)
	if err != nil {
		logError(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		logError(err, "failed to write output proto")
	}
}
