package gocsv

import (
	"reflect"
	"strings"
	"sync"
)

// --------------------------------------------------------------------------
// Reflection helpers

type structInfo struct {
	Fields []fieldInfo
}

// fieldInfo is a struct field that should be mapped to a CSV column, or vica-versa
// Each IndexChain element before the last is the index of an the embedded struct field
// that defines Key as a tag
type fieldInfo struct {
	Key        string
	IndexChain []int
}

var structMap = make(map[reflect.Type]*structInfo)
var structMapMutex sync.RWMutex

func getStructInfo(rType reflect.Type) *structInfo {
	structMapMutex.RLock()
	stInfo, ok := structMap[rType]
	structMapMutex.RUnlock()
	if ok {
		return stInfo
	}
	fieldsList := getFieldInfos(rType, []int{})
	stInfo = &structInfo{fieldsList}
	return stInfo
}

func getFieldInfos(rType reflect.Type, parentIndexChain []int) []fieldInfo {
	fieldsCount := rType.NumField()
	fieldsList := make([]fieldInfo, 0, fieldsCount)
	for i := 0; i < fieldsCount; i++ {
		field := rType.Field(i)
		if field.PkgPath != "" {
			continue
		}
		indexChain := append(parentIndexChain, i)
		// if the field is an embedded struct, create a fieldInfo for each of its fields
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			fieldsList = append(fieldsList, getFieldInfos(field.Type, indexChain)...)
			continue
		}
		fieldTag := field.Tag.Get("csv")
		fieldTags := strings.Split(fieldTag, TagSeparator)
		for _, fieldTagEntry := range fieldTags {
			if fieldTagEntry != "omitempty" {
				fieldTag = fieldTagEntry
				fieldInfo := fieldInfo{IndexChain: indexChain}

				if fieldTag == "-" {
					continue
				} else {
					fieldInfo.Key = fieldTag
				}
				fieldsList = append(fieldsList, fieldInfo)
			}
		}
	}
	return fieldsList
}

func getConcreteContainerInnerType(in reflect.Type) (inInnerWasPointer bool, inInnerType reflect.Type) {
	inInnerType = in.Elem()
	inInnerWasPointer = false
	if inInnerType.Kind() == reflect.Ptr {
		inInnerWasPointer = true
		inInnerType = inInnerType.Elem()
	}
	return inInnerWasPointer, inInnerType
}

func getConcreteReflectValueAndType(in interface{}) (reflect.Value, reflect.Type) {
	value := reflect.ValueOf(in)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	return value, value.Type()
}
