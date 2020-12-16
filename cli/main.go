package main

import (
	"bytes"
	"fmt"
	"github.com/manifoldco/promptui"
	cli "github.com/urfave/cli/v2"
	"html/template"
	"log"
	"os"
	"strings"
)
type ModelDataColumn struct {
	FieldName string
	FieldType string
	ColumnName string
	IsIDType bool
}
type ModelData struct {
	TableName string
	StructName string
	HasID bool
	Columns []ModelDataColumn
}
const modelTpl = `
{{range .Columns }}
type {{.FieldType}} string
func (id {{.FieldType}}) String() string { return string(id) }{{end}}
type {{.StructName}} struct {
{{range .Columns}}	{{.FieldName}} {{.FieldType}} `+"`"+`db:"{{.ColumnName}}"`+"`"+`
{{end}}}
func ({{.StructName}}) TableName() string { return "string"}
func (data *{{.StructName}}) BeforeCreate() {
{{if .HasID }}	if len(data.ID) == 0  { data.ID = ID{{.StructName}}(sq.UUID()) } {{end}}
}
func (User) Column() (col struct {
{{range .Columns}}	{{.FieldName}} sq.Column
{{end}}}) {
{{range .Columns}}	col.{{.FieldName}} = "{{.ColumnName}}"
{{end}}	return
}
`
func RenderModel(data ModelData) string {
	t := template.Must(template.New("renderModel").Parse(modelTpl))
	buffer := bytes.NewBuffer(nil)
	err := t.Execute(buffer, data) ; if err != nil {
		panic(err)
	}
	return buffer.String()
}
func stringFirstWordToUpper(s string) string {
	runes := []rune(s)
	out := strings.ToTitle(string(runes[0])) + string(runes[1:])
	return out
}
func CamelName(name string) string {
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	return strings.Replace(name, " ", "", -1)
}
func ReadModelData() (data ModelData) {
	prompt := promptui.Prompt{
		Label:    "table name",
		Validate: func(s string) error {return nil},
		Pointer: promptui.PipeCursor,
	}
	// TableName
	{
		var err error
		data.TableName, err = prompt.Run() ; if err != nil {
		panic(err)
	}
	}
	// StructName
	{
		data.StructName = strings.TrimSuffix(stringFirstWordToUpper(data.TableName), "s")
	}
	// Columns
	for {
		column := ModelDataColumn{}
		prompt := promptui.Prompt{
			Label:    "column and type (enter END or name string)",
			Validate: func(s string) error {return nil},
			Pointer: promptui.PipeCursor,
		}
		columnAndTypeString, err := prompt.Run() ; if err != nil {
			panic(err)
		}
		columnAndType := strings.Split(columnAndTypeString, " ")
		if len(columnAndType) == 1 {
			columnAndType = append(columnAndType, "string")
			log.Print("auto complete :" , strings.Join(columnAndType, " "))
		}
		columnName := columnAndType[0]
		column.FieldType = columnAndType[1]
		if columnName == "id" || strings.HasSuffix(columnName, "_id") {
			column.IsIDType = true
		}
		if columnName == "" {
			continue
		}
		if columnName == "END" {
			break
		}
		column.ColumnName = columnName
		if columnName == "id" {
			data.HasID = true
		}
		if columnName == "id" {
			column.FieldName = "ID"
		} else {
			column.FieldName = CamelName(stringFirstWordToUpper(columnName))

		}
		data.Columns = append(data.Columns, column)
	}
	data.Columns = append(data.Columns, ModelDataColumn{
		FieldName:  "CreatedAt",
		FieldType:  "time.Time",
		ColumnName: "created_at",
	})
	data.Columns = append(data.Columns, ModelDataColumn{
		FieldName:  "UpdatedAt",
		FieldType:  "time.Time",
		ColumnName: "updated_at",
	})
	return
}
func main() {
	app := &cli.App{
		Name: "goclub/sql cli",
		Usage: "generate model code",
		Action: func(c *cli.Context) error {
			fmt.Println("github.com/goclub/sql/cli")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name: "model",
				Usage:   "generate model",
				Action:  func(c *cli.Context) error {
					log.Print(RenderModel(ReadModelData()))
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
