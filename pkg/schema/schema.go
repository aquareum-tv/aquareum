package schema

import (
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type Schema interface {
	EIP712() (*EIP712SchemaStruct, error)
}

type SchemaStruct struct {
	name    string
	version string
	schema  any
}

func MakeSchema(name, version string, schema any) (Schema, error) {
	return &SchemaStruct{
		name:    name,
		version: version,
		schema:  schema,
	}, nil
}

type EIP712SchemaStruct struct {
	Types      apitypes.Types            `json:"types"`
	Domain     *apitypes.TypedDataDomain `json:"domain"`
	TypeToName map[reflect.Type]string
	NameToType map[string]reflect.Type
}

func (schema *SchemaStruct) EIP712() (*EIP712SchemaStruct, error) {
	var eip712Types = apitypes.Types{
		"EIP712Domain": {
			{
				Name: "name",
				Type: "string",
			},
			{
				Name: "version",
				Type: "string",
			},
		},
	}

	stype := reflect.TypeOf(schema.schema)
	if stype.Kind() != reflect.Struct {
		return nil, fmt.Errorf("schema parameter of MakeEIP712Signer is not a struct")
	}
	fields := reflect.VisibleFields(stype)
	typeToName := map[reflect.Type]string{}
	nameToType := map[string]reflect.Type{}
	for _, field := range fields {
		name := field.Name
		eip712TypeName := fmt.Sprintf("%sData", name)
		if field.Type.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field '%s' in provided schema is not a struct", name)
		}
		typeToName[field.Type] = name
		nameToType[name] = field.Type
		parentType := []apitypes.Type{
			{
				Name: "signer",
				Type: "address",
			},
			{
				Name: "time",
				Type: "int64",
			},
			{
				Name: "data",
				Type: eip712TypeName,
			},
		}
		typeSlice := []apitypes.Type{}

		subfields := reflect.VisibleFields(field.Type)
		for _, subfield := range subfields {
			eipType, err := goToEIP712(subfield)
			if err != nil {
				return nil, fmt.Errorf("error handling type %s: %w", name, err)
			}
			typeSlice = append(typeSlice, eipType)
		}
		eip712Types[name] = parentType
		eip712Types[eip712TypeName] = typeSlice
	}
	return &EIP712SchemaStruct{
		Types: eip712Types,
		Domain: &apitypes.TypedDataDomain{
			Version: schema.version,
			Name:    schema.name,
		},
		TypeToName: typeToName,
		NameToType: nameToType,
	}, nil
}

// turns a go type into an eip712 type
func goToEIP712(field reflect.StructField) (apitypes.Type, error) {
	var typ string
	kind := field.Type.Kind()
	if kind == reflect.String {
		typ = "string"
	} else if kind == reflect.Int64 {
		typ = "int64"
	}
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return apitypes.Type{}, fmt.Errorf("could not find field name for %s", field.Name)
	}
	return apitypes.Type{
		Name: jsonTag,
		Type: typ,
	}, nil
}
