package inputs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestNewProtoFile(t *testing.T) {
	type args struct {
		file *descriptorpb.FileDescriptorProto
	}
	tests := []struct {
		name string
		args args
		want ProtoFile
	}{
		{
			name: "aaa",
			args: args{
				file: &descriptorpb.FileDescriptorProto{
					Name:    proto.String("example"),
					Package: proto.String("examplev1"),
					Options: &descriptorpb.FileOptions{
						GoPackage: proto.String("github.com/carousell/Orion/protoc-gen-orion/internal/testprotos/example;examplev1"),
					},
					Service: []*descriptorpb.ServiceDescriptorProto{
						{
							Name: proto.String("ExampleService"),
							Method: []*descriptorpb.MethodDescriptorProto{
								{
									Name:            proto.String("HttpGet"),
									ClientStreaming: proto.Bool(true),
									ServerStreaming: proto.Bool(true),
								},
							},
						},
					},
					SourceCodeInfo: &descriptorpb.SourceCodeInfo{
						Location: []*descriptorpb.SourceCodeInfo_Location{
							{
								LeadingComments: proto.String("// ORION:URL: GET /example\n"),
								Path:            []int32{6, 1, 2, 1},
							},
						},
					},
				},
			},
			want: ProtoFile{
				Name:          "example",
				PackageName:   "examplev1",
				GoPackagePath: "github.com/carousell/Orion/protoc-gen-orion/internal/testprotos/example",
				Service: []ProtoFileService{
					{
						Name: "ExampleService",
						Method: []ProtoFileServiceMethod{
							{
								Name:            "HttpGet",
								ClientStreaming: true,
								ServerStreaming: true,
							},
						},
					},
				},
				Location: []ProtoFileLocation{
					{
						Comments: "// ORION:URL: GET /example",
						Path:     []int32{6, 1, 2, 1},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewProtoFile(tt.args.file), "NewProtoFile(%v)", tt.args.file)
		})
	}
}

func TestNewProtoParams(t *testing.T) {
	type args struct {
		reqParam string
	}
	tests := []struct {
		name    string
		args    args
		want    ProtoParams
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "default_should_be_false",
			args: args{
				reqParam: "",
			},
			want: ProtoParams{
				StandAloneMode:      false,
				ExportedServiceDesc: false,
			},
			wantErr: assert.NoError,
		},
		{
			name: "enable_standalone_mode_should_be_true",
			args: args{
				reqParam: "standalone-mode=true",
			},
			want: ProtoParams{
				StandAloneMode:      true,
				ExportedServiceDesc: false,
			},
			wantErr: assert.NoError,
		},
		{
			name: "enable_exported_service_desc_should_be_true",
			args: args{
				reqParam: "exported-service-desc=true",
			},
			want: ProtoParams{
				StandAloneMode:      false,
				ExportedServiceDesc: true,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewProtoParams(tt.args.reqParam)
			if !tt.wantErr(t, err, fmt.Sprintf("NewProtoParams(%v)", tt.args.reqParam)) {
				return
			}
			assert.Equalf(t, tt.want, got, "NewProtoParams(%v)", tt.args.reqParam)
		})
	}
}

func Test_parseGoPackage(t *testing.T) {
	type args struct {
		goPackage string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "parse_go_package_should_success",
			args: args{
				goPackage: "github.com/carousell/Orion/protoc-gen-orion/internal/testprotos/example",
			},
			want: "github.com/carousell/Orion/protoc-gen-orion/internal/testprotos/example",
		},
		{
			name: "separate_package_name_should_be_removed",
			args: args{
				goPackage: "github.com/carousell/Orion/protoc-gen-orion/internal/testprotos/example;examplev1",
			},
			want: "github.com/carousell/Orion/protoc-gen-orion/internal/testprotos/example",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, parseGoPackage(tt.args.goPackage), "parseGoPackage(%v)", tt.args.goPackage)
		})
	}
}

func Test_parseBoolValue(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "parse_true_value_should_be_true",
			args: args{
				value: "true",
			},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "parse_false_value_should_be_false",
			args: args{
				value: "false",
			},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "other_value_should_be_error",
			args: args{
				value: "",
			},
			want:    false,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBoolValue(tt.args.value)
			if !tt.wantErr(t, err, fmt.Sprintf("parseBoolValue(%v)", tt.args.value)) {
				return
			}
			assert.Equalf(t, tt.want, got, "parseBoolValue(%v)", tt.args.value)
		})
	}
}
