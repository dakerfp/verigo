package verilog

import (
	"io"
	"text/template"

	"github.com/dakerfp/verigo/meta"
)

var verilogTemplate *template.Template

func init() {
	verilogTemplate = template.Must(template.New("verilog").Parse(`
modulename {{.Name}}
	();

endmodule : {{.Name}}`))
}

func GenerateVerilog(w io.Writer, module meta.Module) error {
	mod := module.Meta()
	return verilogTemplate.Execute(w, mod)
}
