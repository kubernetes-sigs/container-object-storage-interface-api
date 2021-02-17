/*
If this program is on the path of your machine you can invoke it in the following way:

protoc --plugin protoc-gen-gofake --goexample_out=package=cosi,packagePath=sigs.k8s.io/container-object-storage-interface-spec:fake cosi.proto

Requires package, and packagePath to be specified
*/
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type GoFake struct {
	Request    *plugin.CodeGeneratorRequest
	Response   *plugin.CodeGeneratorResponse
	Parameters map[string]string
}

type FakeService struct {
	Name string
	Methods []Method
}

type Method struct {
	Name string
	Input string
	Output string
}

func (runner *GoFake) PrintParameters(w io.Writer) {
	const padding = 3
	tw := tabwriter.NewWriter(w, 0, 0, padding, ' ', tabwriter.TabIndent)
	fmt.Fprintf(tw, "Parameters:\n")
	for k, v := range runner.Parameters {
		fmt.Fprintf(tw, "%s:\t%s\n", k, v)
	}
	fmt.Fprintln(tw, "")
	tw.Flush()
}

func cleanInput(val string) string {
	spl := strings.Split(val, ".")
	return spl[len(spl)-1]
}

func (runner *GoFake) getLocationMessage() map[string][]*FakeService {
	ret := make(map[string][]*FakeService)
	for _, protoFile := range runner.Request.ProtoFile {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", *protoFile.Name)
		fakeServices := make([]*FakeService, 0)
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", protoFile.GetService())
		svcs := protoFile.GetService()
		for _, svc := range svcs {
			_, _ = fmt.Fprintf(os.Stderr, "service: %+v\n", svc)
			current := &FakeService{
				Name:            fmt.Sprintf("%sClient", *svc.Name),
			}
			methods := make([]Method, 0)
			for _, mtd := range svc.Method {
				method := Method{Name: *mtd.Name}
				if mtd.InputType != nil {
					method.Input = cleanInput(*mtd.InputType)
				}
				if mtd.OutputType != nil {
					method.Output = cleanInput(*mtd.OutputType)
				}
				methods = append(methods, method)
			}
			current.Methods = methods

			fakeServices = append(fakeServices, current)
		}
		ret[*protoFile.Name] = fakeServices
	}
	return ret
}

func (runner *GoFake) WriteImports(buf *bytes.Buffer, imports... string) {
	for _, i := range imports {
		buf.WriteString(fmt.Sprintf("\t\"%s\"\n", i))
	}
}

func (runner *GoFake) CreateFakeFile(filename string, fakeSVC []*FakeService) error {
	var outfileName string
	var content string
	outfileName = strings.Replace(filename, ".proto", ".pb.fake.go", -1)
	var mdFile plugin.CodeGeneratorResponse_File
	mdFile.Name = &outfileName
	var buf bytes.Buffer

	pkg := runner.Parameters["package"]
	pkgPath := runner.Parameters["packagePath"]

	buf.WriteString("package fake\n\n")
	buf.WriteString("import (\n")
	runner.WriteImports(&buf, "context", "google.golang.org/grpc")
	buf.WriteString(fmt.Sprintf("\t%s \"%s\"\n", pkg, pkgPath))
	buf.WriteString(")\n\n")
	for _, fakeSVC := range fakeSVC {
		buf.WriteString(fmt.Sprintf("type Fake%s struct {\n", fakeSVC.Name))
		for _, mtd := range fakeSVC.Methods {
			buf.WriteString(fmt.Sprintf("\tFake%s func(ctx context.Context, in *%s.%s, opts ...grpc.CallOption) (*%s.%s, error)\n",
				mtd.Name, pkg, mtd.Input, pkg, mtd.Output))
		}
		buf.WriteString("}\n\n")
		for _, mtd := range fakeSVC.Methods {
			buf.WriteString(fmt.Sprintf("func (f *Fake%s) %s(ctx context.Context, in *%s.%s, opts ...grpc.CallOption) (*%s.%s, error) {\n",
				fakeSVC.Name, mtd.Name, pkg, mtd.Input, pkg, mtd.Output))
			buf.WriteString(fmt.Sprintf("\treturn f.Fake%s(ctx, in, opts...)\n",
				mtd.Name))
			buf.WriteString("}\n\n")
		}
	}
	content = buf.String()
	mdFile.Content = &content
	runner.Response.File = append(runner.Response.File, &mdFile)
	return nil
}

func (runner *GoFake) generateMessageMarkdown() error {
	// This convenience method will return a structure of some types that I use
	for filename, locationMessages := range runner.getLocationMessage() {
		runner.CreateFakeFile(filename, locationMessages)
	}
	return nil
}

func (runner *GoFake) generateCode() error {
	// Initialize the output file slice
	files := make([]*plugin.CodeGeneratorResponse_File, 0)
	runner.Response.File = files

	err := runner.generateMessageMarkdown()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// os.Stdin will contain data which will unmarshal into the following object:
	// https://godoc.org/github.com/golang/protobuf/protoc-gen-go/plugin#CodeGeneratorRequest
	req := &plugin.CodeGeneratorRequest{}
	resp := &plugin.CodeGeneratorResponse{}

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	// You must use the requests unmarshal method to handle this type
	if err := proto.Unmarshal(data, req); err != nil {
		panic(err)
	}

	// You may require more data than what is in the proto files alone.  There are a couple ways in which to do this.
	// The first is by parameters.  Another may be using leading comments in the proto files which I will cover in generateCode.
	parameters := req.GetParameter()
	// =grpc,import_path=mypackage:.
	exampleRunner := &GoFake{
		Request:    req,
		Response:   resp,
		Parameters: make(map[string]string),
	}
	groupkv := strings.Split(parameters, ",")
	for _, element := range groupkv {
		kv := strings.Split(element, "=")
		if len(kv) > 1 {
			exampleRunner.Parameters[kv[0]] = kv[1]
		}
	}
	// Print the parameters for example
	exampleRunner.PrintParameters(os.Stderr)

	err = exampleRunner.generateCode()
	if err != nil {
		panic(err)
	}

	marshalled, err := proto.Marshal(resp)
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(marshalled)
}
