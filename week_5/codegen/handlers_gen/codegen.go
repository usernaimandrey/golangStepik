package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type MethodGenParams struct {
	WraperName   string
	ParamsName   string
	ResultsName  string
	ApiGenParams ApiGenParams
}

type ApiGenParams struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

type ValidateParams struct {
	FiledName          string
	FieldType          string
	NormalizeParamName string
	Validator          interface{}
}

type EnumValidator struct {
	AvailableVals    []string
	AvailableValsStr string
	ParamForError    string
	DefaultVal       string
	TypeValidator    string
	ParamName        string
}

type AnyValidator struct {
	Min           int
	Max           int
	Required      bool
	ForType       string
	ParamName     string
	MinValidate   bool
	MaxValidate   bool
	TypeValidator string
}

type tplSrv struct {
	ApiName string
	Items   []tplSrvItems
}

type tplSrvItems struct {
	Path       string
	WraperName string
}

type tplWraper struct {
	MethodName string
	ParamName  string
	Method     string
	WraperName string
	ApiName    string
	Auth       bool
	Params     []tplWraperItems
}

type tplWraperItems struct {
	FiledName          string
	FieldType          string
	NormalizeParamName string
	ValidatorEnum      EnumValidator
	ValidatorAny       AnyValidator
}

var (
	srvTpl = template.Must(template.New("srvTpl").Parse(`
// {{.ApiName}}
func (srv *{{.ApiName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	{{ range .Items}}
	case "{{.Path}}":
		srv.{{.WraperName}}(w, r)
	{{end}}
	default:
		RenderError(w, http.StatusNotFound, fmt.Errorf("unknown method"))
	}
}
`))

	renderErrorTpl = template.Must(template.New("renderErrorTpl").Parse(`
func RenderError(w http.ResponseWriter, status int, err error) {
	data := map[string]interface{}{
		"error": "",
	}

	switch err.(type) {
	case ApiError:
		errCast := err.(ApiError)
		data["error"] = errCast.Error()
		status = errCast.HTTPStatus
	default:
		data["error"] = err.Error()
	}

	resp, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error"))
	}
	w.WriteHeader(status)
	w.Write(resp)
}
`))

	wraperTpl = template.Must(template.New("wraperTpl").Parse(`
// {{.WraperName}}
func (srv *{{.ApiName}}) {{.WraperName}}(w http.ResponseWriter, r *http.Request) {
{{ if eq .Method "POST" }}
	method := r.Method
	if method != http.MethodPost {
		RenderError(w, http.StatusNotAcceptable, fmt.Errorf("bad method"))
		return
	}
{{ end }}
{{ if .Auth }}
  if r.Header.Get("X-Auth") != "100500" {
		RenderError(w, http.StatusForbidden, fmt.Errorf("unauthorized"))
		return
	}
{{end}}
  params := &{{.ParamName}}{}
{{ range .Params }}
  {{ .NormalizeParamName }} := r.FormValue("{{ .NormalizeParamName }}")
	{{ $validatorEnum := .ValidatorEnum }}
	{{ $validatorAny := .ValidatorAny }}
	{{ if eq .FieldType "int" }}
  var {{ .NormalizeParamName }}Int int
	if len({{ .NormalizeParamName }}) != 0 {
		var err error
		{{ .NormalizeParamName }}Int, err = strconv.Atoi(r.FormValue("{{ .NormalizeParamName }}"))
		if err != nil {
			RenderError(w, http.StatusBadRequest, fmt.Errorf("{{ .NormalizeParamName }} must be int"))
			return
		}
	}
  params.{{ .FiledName }} = {{ .NormalizeParamName }}Int

	{{ else }}
  params.{{ .FiledName }} = {{ .NormalizeParamName }}
	{{ end }}
  {{ if eq $validatorEnum.TypeValidator "enum" }}
	if params.{{ .FiledName }} == "" {
		params.{{ .FiledName }} = "{{ .ValidatorEnum.DefaultVal }}"
	}
	validParam := false
	for _, v := range []{{.FieldType}}{ {{ $validatorEnum.AvailableValsStr }} } {
		if v == params.{{ .FiledName }} {
			validParam = true
		}
	}
	if !validParam {
		err := fmt.Errorf("{{ .NormalizeParamName }} must be one of [{{ $validatorEnum.ParamForError }}]")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
	{{ end }}
	{{ if eq $validatorAny.TypeValidator "any" }}
		{{ if $validatorAny.Required }}
  if len(params.{{ .FiledName }}) == 0 {
	  err := fmt.Errorf("{{ .NormalizeParamName }} must me not empty")
	  RenderError(w, http.StatusBadRequest, err)
	  return
  }
		{{ end }}
		{{ if $validatorAny.MinValidate }}
			{{ if eq $validatorAny.ForType "int" }}
	if params.{{ .FiledName }} < {{ $validatorAny.Min }} {
		err := fmt.Errorf("{{ .NormalizeParamName }} must be >= {{ $validatorAny.Min }}")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
			{{ else }}
	if len(params.{{ .FiledName }}) < {{ $validatorAny.Min }} {
		err := fmt.Errorf("{{ .NormalizeParamName }} len must be >= {{ $validatorAny.Min }}")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
			{{ end }}
		{{ end }}
		{{ if $validatorAny.MaxValidate }}
		  {{ if eq $validatorAny.ForType "int" }}
	if params.{{ .FiledName }} > {{ $validatorAny.Max }} {
		err := fmt.Errorf("{{ .NormalizeParamName }} must be <= {{ $validatorAny.Max }}")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
			{{ else }}
	if len(params.{{ .FiledName }}) > {{ $validatorAny.Max }} {
		err := fmt.Errorf("{{ .NormalizeParamName }} len must be <= {{ $validatorAny.Max }}")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
			{{ end }}
		{{ end }}
	{{ end }}
{{ end }}
  user, err := srv.{{ .MethodName }}(r.Context(), *params)
	if err != nil {
		RenderError(w, http.StatusInternalServerError, err)
		return
	}

	resp := map[string]interface{}{
		"error":    "",
		"response": user,
	}

	data, err := json.Marshal(resp)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{error:" + err.Error() + "}"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
`))
)

func EnumValidatorBuilder(params string) EnumValidator {
	data := strings.Split(params, ",")
	validator := EnumValidator{}
	validator.TypeValidator = "enum"
	for _, v := range data {
		if strings.Contains(v, "enum") {
			cut, ok := strings.CutPrefix(v, "enum=")
			if ok {
				validator.AvailableVals = strings.Split(cut, "|")
				validator.ParamForError = strings.Join(strings.Split(cut, "|"), ", ")
				for i, str := range strings.Split(cut, "|") {
					sep := ", "
					if len(strings.Split(cut, "|"))-1 == i {
						sep = ""
					}
					strN := fmt.Sprintf(`"%v"%v `, str, sep)
					validator.AvailableValsStr += strN
				}
			}
		} else if strings.Contains(v, "default") {
			validator.DefaultVal = strings.Split(v, "=")[1]
		}
	}
	return validator
}

func AnyValidatorBuilder(parms string, typeParam string) AnyValidator {
	validator := AnyValidator{}
	validator.TypeValidator = "any"
	validator.Required = strings.Contains(parms, "required")
	validator.ForType = typeParam

	data := strings.Split(parms, ",")
	for _, v := range data {
		if strings.Contains(v, "required") {
			continue
		}

		splitedParam := strings.Split(v, "=")
		param := splitedParam[0]
		val := splitedParam[1]
		if param == "max" {
			maxInt, err := strconv.Atoi(val)

			if err != nil {
				panic(err)
			}

			validator.Max = maxInt
			validator.MaxValidate = true
		} else if param == "min" {
			minInt, err := strconv.Atoi(val)

			if err != nil {
				panic(err)
			}
			validator.Min = minInt
			validator.MinValidate = true
		} else if param == "paramname" {
			validator.ParamName = val
		} else {
			err := fmt.Errorf("unknow param name %s", param)
			panic(err)
		}
	}
	return validator
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	packages := []string{"encoding/json", "fmt", "net/http", "strconv"}

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `import (`)
	for _, pack := range packages {
		fmt.Fprintln(out, "	"+`"`+pack+`"`)
	}
	fmt.Fprintln(out, `)`)
	fmt.Fprintln(out)
	renderErrorTpl.Execute(out, "")
	fmt.Fprintln(out)

	generateParams := make(map[string]map[string]MethodGenParams)
	validateParams := make(map[string][]ValidateParams)
	responseParams := make(map[string]map[string]string)

	for _, f := range node.Decls {
		funcDec, okFuncDecl := f.(*ast.FuncDecl)
		genDecl, okGenDecl := f.(*ast.GenDecl)

		if okFuncDecl {
			FuncDeclCollecterParams(funcDec, generateParams)
		} else if okGenDecl {
			GenDeclCollecterParams(genDecl, validateParams, responseParams)
		} else {
			continue
		}
	}

	for api, apiParams := range generateParams {
		srv := tplSrv{}
		srvItems := []tplSrvItems{}
		srv.ApiName = api
		for method, apiParam := range apiParams {
			srvItem := tplSrvItems{
				Path:       apiParam.ApiGenParams.Url,
				WraperName: apiParam.WraperName,
			}
			srvItems = append(srvItems, srvItem)
			validators := validateParams[apiParam.ParamsName]
			normalizeValidators := make([]tplWraperItems, 0)

			for _, validator := range validators {
				tmlItem := tplWraperItems{}
				tmlItem.FieldType = validator.FieldType
				tmlItem.FiledName = validator.FiledName
				tmlItem.NormalizeParamName = validator.NormalizeParamName

				switch validator.Validator.(type) {
				case EnumValidator:
					v := validator.Validator.(EnumValidator)
					tmlItem.ValidatorEnum = v
				case AnyValidator:
					v := validator.Validator.(AnyValidator)
					tmlItem.ValidatorAny = v
				default:
					panic(fmt.Errorf("unknow validator"))
				}
				normalizeValidators = append(normalizeValidators, tmlItem)
			}

			tplWraper := tplWraper{
				method,
				apiParam.ParamsName,
				apiParam.ApiGenParams.Method,
				apiParam.WraperName,
				api,
				apiParam.ApiGenParams.Auth,
				normalizeValidators,
			}
			wraperTpl.Execute(out, tplWraper)
		}
		srv.Items = srvItems
		srvTpl.Execute(out, srv)

	}
}

func GenDeclCollecterParams(genDecl *ast.GenDecl, validateParams map[string][]ValidateParams, responseParams map[string]map[string]string) {
	for _, spec := range genDecl.Specs {
		currType, ok := spec.(*ast.TypeSpec)
		if !ok {
			fmt.Printf("SKIP %#T is not ast.TypeSpec\n", spec)
			continue
		}

		currStruct, ok := currType.Type.(*ast.StructType)

		if !ok {
			fmt.Printf("SKIP %#T is not ast.StructType\n", currStruct)
			continue
		}

		typeName := currType.Name.Name

		fmt.Println(currType.Name.Name, "this is type")

		for _, field := range currStruct.Fields.List {
			if field.Tag == nil {
				continue
			}

			tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])

			if tag.Get("json") != "" {
				continue
			}

			if tag.Get("apivalidator") == "" {
				continue
			}

			fieldType := field.Type.(*ast.Ident).Name
			tagVal := tag.Get("apivalidator")

			var validator interface{}

			if strings.Contains(tagVal, "enum") {
				validator = EnumValidatorBuilder(tagVal)
			} else {
				validator = AnyValidatorBuilder(tagVal, fieldType)
			}

			fieldName := field.Names[0].Name
			valParams := &ValidateParams{}
			valParams.FieldType = fieldType
			valParams.FiledName = fieldName
			valParams.Validator = validator

			switch validator.(type) {
			case AnyValidator:
				validatorCats := validator.(AnyValidator)
				if validatorCats.ParamName == "" {
					valParams.NormalizeParamName = strings.ToLower(fieldName)
				} else {
					valParams.NormalizeParamName = validatorCats.ParamName
				}
			case EnumValidator:
				validatorCats := validator.(EnumValidator)
				if validatorCats.ParamName == "" {
					valParams.NormalizeParamName = strings.ToLower(fieldName)
				} else {
					valParams.NormalizeParamName = validatorCats.ParamName
				}
			default:
				panic(fmt.Errorf("unknow validator"))
			}

			if _, exists := validateParams[typeName]; exists {
				validateParams[typeName] = append(validateParams[typeName], *valParams)
			} else {
				validateParams[typeName] = make([]ValidateParams, 0)
				validateParams[typeName] = append(validateParams[typeName], *valParams)
			}

		}
	}
}

func FuncDeclCollecterParams(funcDec *ast.FuncDecl, generateParams map[string]map[string]MethodGenParams) {
	if funcDec.Doc == nil {
		return
	}

	rec := funcDec.Recv.List[0]
	switch expr := rec.Type.(type) {
	case *ast.StarExpr:
		if ident, ok := expr.X.(*ast.Ident); ok {
			funcName := funcDec.Name.Name
			structName := ident.Name
			genParam := &MethodGenParams{
				WraperName: structName + funcName + "Wraper",
			}

			needCodegen := false
			for _, comment := range funcDec.Doc.List {
				s := "// apigen:api"
				needCodegen = needCodegen || strings.HasPrefix(comment.Text, s)
				if !needCodegen {
					continue
				}
				normalizeParam, _ := strings.CutPrefix(comment.Text, s)
				apiGenParams := &ApiGenParams{}
				json.Unmarshal([]byte(normalizeParam), apiGenParams)
				genParam.ApiGenParams = *apiGenParams

			}

			if needCodegen {
				param := funcDec.Type.Params.List[1]
				result := funcDec.Type.Results.List[0]
				paramType := fmt.Sprintf("%s", param.Type)
				resultType := fmt.Sprintf("%s", result.Type)
				resultType = strings.Replace(strings.Split(resultType, " ")[1], "}", "", -1)
				genParam.ParamsName = paramType
				genParam.ResultsName = resultType

				if _, exists := generateParams[structName]; exists {
					generateParams[structName][funcName] = *genParam
				} else {
					generateParams[structName] = map[string]MethodGenParams{
						funcName: *genParam,
					}
				}
			}
		}
	}
}
