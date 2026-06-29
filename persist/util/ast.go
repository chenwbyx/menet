package util

import (
	"log"
	"github.com/fatih/structtag"
	"go/ast"
	"strconv"
)

func GetFieldTag(field *ast.Field, key string) *structtag.Tag {
	if field.Tag == nil {
		return &structtag.Tag{}
	}

	s, _ := strconv.Unquote(field.Tag.Value)
	tags, err := structtag.Parse(s)
	if err != nil {
		log.Printf("parse tag string:%s failed:%v", field.Tag.Value, err)
		return &structtag.Tag{}
	}
	tag, err := tags.Get(key)
	if err != nil {
		return &structtag.Tag{}
	}

	return tag
}

func GetFieldTags(field *ast.Field) *structtag.Tags {
	if field.Tag == nil {
		return &structtag.Tags{}
	}

	s, _ := strconv.Unquote(field.Tag.Value)
	tags, err := structtag.Parse(s)
	if err != nil {
		log.Printf("parse tag string:%s failed:%v", field.Tag.Value, err)
		return &structtag.Tags{}
	}
	return tags
}

func GetFieldName(field *ast.Field) string {
	if len(field.Names) > 0 {
		return field.Names[0].Name
	}

	return ""
}

func GetFieldType(field *ast.Field) string {
	if v, ok := field.Type.(*ast.Ident); ok {
		return v.Name
	}
	return ""
}

func GetStructBuildInTypeFields(node ast.Node) []*ast.Field {
	var fields []*ast.Field
	nodeType, ok := node.(*ast.StructType)
	if !ok {
		return nil
	}
	for _, field := range nodeType.Fields.List {
		if _, ok := field.Type.(*ast.Ident); !ok {
			continue
		}
		if field.Type.(*ast.Ident).Obj != nil {
			continue
		}
		fields = append(fields, field)
	}

	return fields
}

func GetStructBuildOutTypeFields(node ast.Node) []*ast.Field {
	var fields []*ast.Field
	nodeType, ok := node.(*ast.StructType)
	if !ok {
		return nil
	}
	for _, field := range nodeType.Fields.List {
		if _, ok := field.Type.(*ast.Ident); ok {
			continue
		}
		//if field.Type.(*ast.Ident).Obj == nil {
		//	continue
		//}
		fields = append(fields, field)
	}

	return fields
}

func getFiledTypeRecursion(fieldType interface{}) (typeStr string) {
	switch t := fieldType.(type) {
	case *ast.MapType:
		// Key
		typeStr += "map["
		typeStr += getFiledTypeRecursion(t.Key)
		typeStr += "]"
		// Value
		typeStr += getFiledTypeRecursion(t.Value)
	case *ast.ArrayType:
		// Slice or Array
		typeStr += "["
		if t.Len != nil {
			typeStr += t.Len.(*ast.BasicLit).Value
		}
		typeStr += "]"
		// Value
		typeStr += getFiledTypeRecursion(t.Elt)
	case *ast.Ident:
		typeStr += t.Name
	}

	return
}

func getFiledCopyRecursion(deep int, filedName string, fieldType interface{}) (blockStr string) {
	deepOldStr := strconv.Itoa(deep-1)
	deepStr := strconv.Itoa(deep)
	if deep == 0 {
		switch t := fieldType.(type) {
		case *ast.MapType:
			blockStr += `
    for k0, v0 := range src.` + filedName + "{"

			blockStr += getFiledCopyRecursion(deep+1, filedName+"[k"+deepStr+"]", t.Value)

			blockStr += `
    }`
		case *ast.ArrayType:
			// Slice or Array
			//typeStr += "["
			//if t.Len != nil {
			//	typeStr += t.Len.(*ast.BasicLit).Value
			//}
			//typeStr += "]"
			//// Value
			//typeStr += getFiledTypeRecursion(t.Elt)
		case *ast.Ident:
			// 不会执行到这里
			blockStr += `
    dst.` + filedName + " = src." + filedName
		}

	} else {
		switch t := fieldType.(type) {
		case *ast.MapType:
			blockStr += `
    for k`+deepStr +`, v`+deepStr+` := range ` + "v"+deepOldStr + "{"

			blockStr += getFiledCopyRecursion(deep+1, filedName+"[k"+deepStr+"]", t.Value)

			blockStr += `
    }`
		case *ast.ArrayType:
			// Slice or Array
			//typeStr += "["
			//if t.Len != nil {
			//	typeStr += t.Len.(*ast.BasicLit).Value
			//}
			//typeStr += "]"
			//// Value
			//typeStr += getFiledTypeRecursion(t.Elt)
		case *ast.Ident:
			blockStr += `
`
		}

	}

	return
}

func GetFiledDeepCopyString(field *ast.Field) (block string) {
	fieldName := GetFieldName(field)
	fieldType := getFiledTypeRecursion(field.Type)
	log.Println("@@@@@@@@dst = src@@@@@@@@@", fieldType)
	block += `
    dst.` + fieldName + " = make(" + fieldType + ")"

	block += getFiledCopyRecursion(0, fieldName, field.Type)

	return
}
