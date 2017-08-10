package verilog

import (
	"io"
	"text/template"

	"github.com/dakerfp/verigo/meta"
)

var verilogTemplate *template.Template

func init() {

	verilogTemplate = template.Must(template.New("verilog").Parse(`
{{- define "inports"}}
	{{- range $i, $n := $}}
		{{- if $i}}	 {{end}}input {{$n.T}} {{$n.Name}}
		{{- ",\n"}}
	{{- end}}
{{- end}}
{{- define "outports"}}
	{{- range $i, $n := $}}
		{{- if gt $i 0}}	 {{end}}output {{$n.T}} {{$n.Name}}
		{{- ",\n"}}
	{{- end}}
{{- end}}
modulename {{.Name}}
	({{template "inports" .Inputs}}
	 {{template "outports" .Outputs}}	);

endmodule : {{.Name}}
`))
}

func GenerateVerilog(w io.Writer, module meta.Module) error {
	mod := module.Meta()
	return verilogTemplate.Execute(w, mod)
}
