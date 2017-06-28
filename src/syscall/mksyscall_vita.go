package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-clang/v3.8/clang"
	"strings"
	"errors"
)

var reservedList = []string{
	"break",
	"default",
	"func",
	"interface",
	"select",
	"case",
	"defer",
	"go",
	"map",
	"struct",
	"chan",
	"else",
	"goto",
	"package",
	"switch",
	"const",
	"fallthrough",
	"if",
	"range",
	"type",
	"continue",
	"for",
	"import",
	"return",
	"var",
}

type funcArg struct {
	argName string
	argType clang.Type
}

type argArray []funcArg

type funcInfo struct {
	name        string
	args        argArray
	returnValue clang.Type
	variadic    bool
}

type enumEntry struct {
	name  string
	value uint64
}

type enumArray []enumEntry

type enumInfo struct {
	name    string
	entries enumArray
}

type structEntry struct {
	name    string
	argType clang.Type
}

type structInfo struct {
	name   string
	fields []structEntry
	union  bool
	nested bool
}

type structArray []structInfo

type typedefInfo struct {
	name    string
	argType clang.Type
}

type typedefArray []typedefInfo

var fname = flag.String("fname", "", "the file to analyze")

func main() {
	os.Exit(cmd(os.Args[1:]))
}

func stringInSlice(s string, list []string) bool {
	for _, b := range list {
		if b == s {
			return true
		}
	}
	return false
}

func (t typedefInfo) hasScePrefix() bool {
	return strings.HasPrefix(t.name, "Sce") ||
		strings.HasPrefix(t.name, "_sce")
}

func validStructName(name string) bool {
	return strings.HasPrefix(name, "Sce") ||
		strings.HasPrefix(name, "_sce") ||
		strings.HasPrefix(name, "MusicExportParam") ||
		strings.HasPrefix(name, "PhotoExportParam") ||
		strings.HasPrefix(name, "ScreenShotParam")
}

func cmd(args []string) int {
	if err := flag.CommandLine.Parse(args); err != nil {
		fmt.Printf("ERROR: %s", err)

		return 1
	}

	if *fname == "" {
		flag.Usage()
		fmt.Printf("please provide a file name to analyze\n")

		return 1
	}

	idx := clang.NewIndex(0, 1)
	defer idx.Dispose()

	tuArgs := []string{}
	if len(flag.Args()) > 0 && flag.Args()[0] == "-" {
		tuArgs = make([]string, len(flag.Args()[1:]))
		copy(tuArgs, flag.Args()[1:])
	}

	tu := idx.ParseTranslationUnit(*fname, tuArgs, nil, 0)
	defer tu.Dispose()

	diagnostics := tu.Diagnostics()
	for _, d := range diagnostics {
		fmt.Fprintf(os.Stderr, "PROBLEM: %s", d.Spelling())
	}

	cursor := tu.TranslationUnitCursor()

	var funcList []funcInfo
	var enumList []enumInfo
	var structList []structInfo
	var typedefList []typedefInfo

	var typeData typedefInfo
	var cursorInfo clang.Cursor

	enumIndex := -1
	structIndex := -1

	cursor.Visit(func(cursor, parent clang.Cursor) clang.ChildVisitResult {
		if cursor.IsNull() {
			fmt.Printf("cursor: <none>\n")

			return clang.ChildVisit_Continue
		}

		//fmt.Printf("%s: %s (%s)\n", cursor.Kind().Spelling(), cursor.Spelling(), cursor.USR())

		switch cursor.Kind() {
		case clang.Cursor_TypedefDecl:
			typeData.name = cursor.Spelling()
			// get the underlying type
			typeData.argType = cursor.TypedefDeclUnderlyingType()

			// check if the typedef is from a struct or enum
			switch typeData.argType.CanonicalType().Kind() {
			case clang.Type_Record:
				if len(structList) > 0 && structList[structIndex].name == "" {
					structList[structIndex].name = typeData.name
				}
			case clang.Type_Enum:
				if len(enumList) > 0 && enumList[enumIndex].name == "" {
					enumList[enumIndex].name = typeData.name
				}
			}

			if typeData.hasScePrefix() {
				switch typeData.argType.Kind() {
				case clang.Type_Unexposed:
					switch typeData.argType.CanonicalType().Kind() {
					case clang.Type_FunctionProto:
						typedefList = append(typedefList, typedefInfo{
							name:    typeData.name,
							argType: typeData.argType.CanonicalType(),
						})
					}
				default:
					typedefList = append(typedefList, typedefInfo{
						name:    typeData.name,
						argType: cursor.TypedefDeclUnderlyingType().CanonicalType(),
					})
				}
			}
		case clang.Cursor_EnumDecl:
			enumList = append(enumList, enumInfo{
				name: cursor.Spelling(),
			})
			enumIndex++

			// follow the enum
			return clang.ChildVisit_Recurse
		case clang.Cursor_EnumConstantDecl:
			// add an entry to the current enum
			enumList[enumIndex].entries = append(enumList[enumIndex].entries, enumEntry{
				name:  cursor.Spelling(),
				value: cursor.EnumConstantDeclUnsignedValue(),
			})
		case clang.Cursor_StructDecl:
			// check if nested struct
			if parent.Equal(cursorInfo) {
				structList[structIndex].nested = true
				return clang.ChildVisit_Continue
			}
			cursorInfo = cursor
			structName := cursor.Spelling()

			structList = append(structList, structInfo{
				name: structName,
			})
			structIndex++

			// follow the struct
			return clang.ChildVisit_Recurse
		case clang.Cursor_UnionDecl:
			// create union as struct
			structList = append(structList, structInfo{
				name: cursor.Spelling(),
				// mark as union
				union: true,
			})
			structIndex++

			// follow the enum
			return clang.ChildVisit_Recurse
		case clang.Cursor_FieldDecl:
			// add field to struct/union
			structList[structIndex].fields = append(structList[structIndex].fields, structEntry{
				name:    cursor.Spelling(),
				argType: cursor.Type(),
			})
		case clang.Cursor_ClassDecl, clang.Cursor_Namespace:
			// unused, C header only
			return clang.ChildVisit_Recurse
		case clang.Cursor_FunctionDecl:
			funcName := cursor.Spelling()
			returnType := cursor.ResultType()
			argc := int(cursor.NumArguments())
			argList := make([]funcArg, argc)

			for i := range argList {
				arg := cursor.Argument(uint32(i))
				argList[i].argName = arg.Spelling()
				argList[i].argType = arg.Type()
			}

			f := funcInfo{
				name:        funcName,
				args:        argList,
				returnValue: returnType,
				variadic:    cursor.IsVariadic(),
			}

			funcList = append(funcList, f)
		}

		return clang.ChildVisit_Continue
	})

	if len(diagnostics) > 0 {
		fmt.Fprintln(os.Stderr, "NOTE: There were problems while analyzing the given file")
	}

	fmt.Printf("package vita\n\n")
	printEnums(enumList)
	fmt.Printf("\n\n")
	printTypedefs(typedefList)
	fmt.Printf("\n\n")
	printStructs(structList)
	fmt.Printf("\n\n")
	printCgo(funcList)
	fmt.Printf("\n\n")
	printVars(funcList)
	fmt.Printf("\n\n")
	printFuncs(funcList)

	return 0
}

func getGoType(t clang.Type, isFunction bool) (string, error) {
	switch t.Kind() {
	case clang.Type_Bool:
		return "bool", nil
	case clang.Type_Int:
		return "int", nil
	case clang.Type_Char_S:
		return "byte", nil
	case clang.Type_SChar:
		return "int8", nil
	case clang.Type_Void:
		return "uintptr", nil
	case clang.Type_UInt:
		return "uint", nil
	case clang.Type_UChar:
		return "uint8", nil
	case clang.Type_Float:
		return "float32", nil
	case clang.Type_UShort:
		return "uint16", nil
	case clang.Type_Short:
		return "int16", nil
	case clang.Type_ULongLong:
		return "uint64", nil
	case clang.Type_LongLong:
		return "int64", nil
	case clang.Type_ULong:
		return "uint32", nil
	case clang.Type_Long:
		return "int32", nil
	case clang.Type_Double:
		return "float64", nil
	case clang.Type_Unexposed:
		return "uintptr /* XXX */", nil
	case clang.Type_Enum:
		if strings.HasPrefix(t.Spelling(), "enum ") {
			return t.Spelling()[5:] + "_E /* Enum */", nil
		} else {
			return t.Spelling() + "_E /* Enum */", nil
		}
	case clang.Type_FunctionProto:
		return "uintptr /* function */", nil
	case clang.Type_IncompleteArray:
		arrayType := t.ArrayElementType()
		switch arrayType.Kind() {
		case clang.Type_Char_S:
			if isFunction {
				return "string", nil
			} else {
				return "*Char", nil
			}
		case clang.Type_UChar:
			return "*byte", nil
		case clang.Type_Pointer:
			r, err := getGoType(t.ArrayElementType(), isFunction)
			if err != nil {
				return "", err
			}
			return "*" + r, nil
		case clang.Type_Typedef:
			size := t.ArraySize()
			r, err := getGoType(arrayType.CanonicalType(), isFunction)
			if err != nil {
				return "", err
			}
			field := fmt.Sprintf("[%v]%s", size, r)
			return field, nil
		default:
			size := t.ArraySize()
			r, err := getGoType(arrayType.CanonicalType(), isFunction)
			if err != nil {
				return "", err
			}
			field := fmt.Sprintf("[%v]%s", size, r)
			return field, nil
		}
	case clang.Type_ConstantArray:
		size := t.ArraySize()
		arrayType := t.ArrayElementType()
		switch arrayType.Kind() {
		case clang.Type_Char_S:
			return fmt.Sprintf("[%v]Char", size), nil
		case clang.Type_UChar:
			return fmt.Sprintf("[%v]byte", size), nil
		case clang.Type_Pointer:
			r, err := getGoType(t.ArrayElementType(), isFunction)
			if err != nil {
				return "", err
			}
			return "*" + r, nil
		case clang.Type_Typedef:
			r, err := getGoType(arrayType.CanonicalType(), isFunction)
			if err != nil {
				return "", err
			}
			field := fmt.Sprintf("[%v]%s", size, r)
			return field, nil
		default:
			size := t.ArraySize()
			r, err := getGoType(arrayType.CanonicalType(), isFunction)
			if err != nil {
				return "", err
			}
			field := fmt.Sprintf("[%v]%s", size, r)
			return field, nil
		}
	case clang.Type_Pointer:
		pointee := t.PointeeType()
		switch pointee.Kind() {
		case clang.Type_Void:
			return "uintptr", nil
		case clang.Type_FunctionProto:
			return "uintptr /* function */", nil
		case clang.Type_Char_S:
			if pointee.IsConstQualifiedType() {
				if isFunction {
					return "string", nil
				} else {
					return "*Char", nil
				}
			} else {
				return "*Char", nil
			}
		case clang.Type_Unexposed:
			return "uintptr", nil
		case clang.Type_Record:
			if strings.HasPrefix(t.Spelling(), "struct ") {
				return pointee.Spelling()[7:] + "_S", nil
			} else {
				return pointee.Spelling() + "_S", nil
			}
		case clang.Type_Typedef:
			cr, err := getGoType(pointee.CanonicalType(), isFunction)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("*%s", cr), nil
		}

		r, err := getGoType(pointee, isFunction)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("*%s", r), nil

	case clang.Type_Typedef:
		switch t.CanonicalType().Kind() {
		case clang.Type_Record:
			return t.Spelling() + "_S", nil
		case clang.Type_Enum:
			return t.Spelling() + "_E", nil
		}

		if !validStructName(t.Spelling()) {
			return getGoType(t.CanonicalType(), isFunction)
		}

		return t.Spelling(), nil
	case clang.Type_Record:
		if strings.HasPrefix(t.Spelling(), "struct ") {
			return t.Spelling()[7:] + "_S", nil
		} else if strings.HasPrefix(t.Spelling(), "union ") {
			return t.Spelling()[6:] + "_S", nil
		} else if strings.HasPrefix(t.Spelling(), "const struct ") {
			return t.Spelling()[13:] + "_S", nil
		} else if strings.HasPrefix(t.Spelling(), "const ") {
			return t.Spelling()[6:] + "_S", nil
		} else {
			return t.Spelling() + "_S", nil
		}
	}

	return "", errors.New(fmt.Sprintf("Unknown type: %v (%v)", t.Spelling(), t.Kind()))
}

func (t argArray) parseArgs(variadic bool) (string, error) {
	var result string
	for i, e := range t {

		var argName string
		if e.argName == "" {
			// no arg name, assign a generic one
			argName = fmt.Sprintf("arg%v", i)
		} else {
			if stringInSlice(e.argName, reservedList) {
				// if the name is reserved then append an underscore
				argName = e.argName + "_"
			} else {
				argName = e.argName
			}
		}

		r, err := getGoType(e.argType, true)
		if err != nil {
			return "", err
		}

		if strings.HasPrefix(r, "*_sce") {
			r = "*Sce" + r[5:]
		}

		result += argName + " " + r

		if i < len(t)-1 {
			result += ", "
		}
	}

	if variadic {
		if len(t) > 0 {
			result += ", "
		}
		result += fmt.Sprintf("arg%v... string", len(t))
	}

	return result, nil
}

// first letter to uppercase, remove underscore
func toExportableGoName(cname string) string {
	if strings.HasPrefix(cname, "_") {
		return strings.ToUpper(cname[1:2]) + cname[2:]
	} else {
		return strings.ToUpper(cname[0:1]) + cname[1:]
	}
}

func printStructs(structs []structInfo) {
	for _, entry := range structs {
		if !validStructName(entry.name) {
			continue
		}

		structName := toExportableGoName(entry.name) + "_S"

		if entry.union {
			fmt.Println("/* union */")
			fmt.Printf("type %s struct {\n", structName)
		} else {
			if entry.nested {
				fmt.Println("/* unsupported: nested struct */")
				fmt.Printf("type %s struct {}\n", structName)
				continue
			} else {
				fmt.Printf("type %s struct {\n", structName)
			}
		}
		for _, field := range entry.fields {
			fieldName := toExportableGoName(field.name)
			if fieldName == "PInfo" {
				fmt.Println("")
			}
			fieldType, err := getGoType(field.argType, false)
			if err != nil {
				fmt.Print(err)
				os.Exit(1)
			}
			fmt.Printf("\t%s %s\n", fieldName, fieldType)

		}
		fmt.Printf("}\n\n")
	}
}

func printCgo(funcs []funcInfo) {
	for _, entry := range funcs {
		if strings.HasPrefix(entry.name, "sce") || strings.HasPrefix(entry.name, "_sce") {
			fmt.Printf("//go:cgo_import_static %s\n", entry.name)
		}
	}
	fmt.Printf("\n")
	for _, entry := range funcs {
		if strings.HasPrefix(entry.name, "sce") || strings.HasPrefix(entry.name, "_sce") {
			fmt.Printf("//go:linkname %s %s\n", toExportableGoName(entry.name), entry.name)
		}
	}
}

func printVars(funcs []funcInfo) {
	fmt.Printf("type libFunc uintptr\n\n")
	fmt.Printf("var (\n")
	for i, entry := range funcs {
		if strings.HasPrefix(entry.name, "sce") || strings.HasPrefix(entry.name, "_sce") {
			if i < len(funcs)-1 {
				fmt.Printf("\t lib_%s,\n", toExportableGoName(entry.name))
			} else {
				fmt.Printf("\t lib_%s libFunc\n", toExportableGoName(entry.name))
			}
		}
	}
	fmt.Printf(")\n")
}

func printFuncs(funcs []funcInfo) {
	total := 0
	for _, entry := range funcs {

		var returnStr string
		if entry.returnValue.Kind() == clang.Type_Void {
			returnStr = ""
		} else {
			r, err := getGoType(entry.returnValue, true)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			returnStr = " (rv " + r + ")"
		}

		if strings.HasPrefix(entry.name, "sce") || strings.HasPrefix(entry.name, "_sce") {
			total++
			args, err := entry.args.parseArgs(entry.variadic)
			funcName := toExportableGoName(entry.name)
			if err != nil {
				fmt.Printf("// func %s(interface{})%s\n", funcName, returnStr)
			} else {
				fmt.Printf("func %s(%s)%s {\n", funcName, args, returnStr)
				if returnStr != "" {
					if strings.HasPrefix(returnStr, " (rv *") {
						fmt.Printf("\t return nil\n}\n")
					} else if strings.HasPrefix(returnStr, " (rv string") {
						fmt.Printf("\t return \"\"\n}\n")
					} else {
						fmt.Printf("\t return 0\n}\n")
					}
				} else {
					fmt.Printf("\n}\n")
				}
			}
		}
	}
}

func printEnums(enums []enumInfo) {
	for _, e := range enums {
		var enumName string
		if e.name != "" {
			enumName = e.name + "_E"
			fmt.Printf("type %s uint\n\n", enumName)
		} else {
			enumName = "uint"
		}

		fmt.Println("const (")
		for i, a := range e.entries {
			if i == 0 {
				fmt.Printf("\t %s %s = 0x%08X\n", a.name, enumName, a.value)
			} else {
				fmt.Printf("\t %s = 0x%08X\n", a.name, a.value)
			}
		}

		fmt.Print(")\n\n")
	}
}

func printTypedefs(defs []typedefInfo) {
	fmt.Printf("type Char byte\n\n")
	for _, e := range defs {
		typeName := toExportableGoName(e.name)
		typeRef, err := getGoType(e.argType, false)
		if err != nil {
			fmt.Printf("// type %s %s\n", typeName, e.argType.Spelling())
		} else {
			if e.argType.Kind() == clang.Type_Pointer &&
				e.argType.PointeeType().Kind() == clang.Type_FunctionProto {
				fmt.Printf("// %s\n", e.argType.Spelling())
			}
			fmt.Printf("type %s %s\n", typeName, typeRef)
		}
	}
}
