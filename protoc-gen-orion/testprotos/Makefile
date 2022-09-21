# Generate the *.orion.pb.go into each folder.
# The generated files shouldn't be changed, unless you change the orion template format.
# Notice: Before running the script, you should install the testing binary of protoc-gen-orion in your device.
all:
# The default protoc-gen-orion will provide the unexported-service-desc.
# It's suitable for protoc-gen-go version lower and equal than v1.5.2
	protoc --orion_out="." example/example.proto

# For protoc-gen-go version higher and equal than v1.20.0, please enable the exported-service-desc option.
	protoc --orion_out="exported-service-desc=true:." exported_service_desc/example.proto

# Enable the standalone-mode option will allow your to place *.orion.pb.go and *.pb.go into different folder.
# The *.orion.pb.go imports the *.pb.go by go_package defined in *.proto.
	protoc --orion_out="standalone-mode=true:." standalone_mode/example.proto
	mv standalone_mode/example.proto.orion.pb.go standalone_mode/orion/