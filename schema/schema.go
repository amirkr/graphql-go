package schema

import (
	"log"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"

	// "github.com/amirkr/graphql-go/model"
	"time"

	"github.com/amirkr/graphql-go/resolver"
	"github.com/graphql-go/graphql"
	"github.com/relvacode/iso8601"
)

func GetSchema() graphql.Schema {
	// Schema
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
		"author": GenerateAuthorConfig(),
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	return schema
}

func GenerateAuthorConfig() (objectConfig *graphql.Field) {
	// sysdataSchema := `
	// #sysdata: {
	// 	lkwd_id: int | *0
	// 	created_at: string | *"2017-32-32T07:20:50.52Z"
	// 	isDelete: bool| *false
	// 	price: float | *0.0
	// 	name: string | *""
	// 	databaseType: string | *""
	// 	exportType: string | *""
	// 	exportCategory: string | *""
	// 	{
	// 		if sysdata.databaseType == "apartment" {
	// 			exportType:     "Areas"
	// 			exportCategory: "Apartment"
	// 		}
	// 		if sysdata.databaseType == "public_space" {
	// 			exportType:     "Areas"
	// 			exportCategory: "Apartment"
	// 		}
	// 		if sysdata.databaseType == "animation" {
	// 			exportType:     "Avakin"
	// 			exportCategory: "animation"
	// 		}
	// 	}
	// 	object: {
	// 		obj_id: int | *0
	// 		obj_name: string | *""
	// 	}
	// 	...
	// }
	// sysdata: #sysdata
	// `

	// cueSchemaStr := sysdataSchema + `
	cueSchemaStr := `
	#author: {
		id: string | *""
		firstname: string  | *""
		lastname: string | *""
		createdat: string | *"2017-10-08T07:20:50.52Z"
		object: {
			obj_id: int | *0
			obj_name: string | *""
		}
		editorsid: [string] | *["subzero","zero"]
	}

	author: #author
	`

	cueCtx := cuecontext.New()
	cueSchema := cueCtx.CompileString(cueSchemaStr)
	schemaFields, err := cueSchema.Fields()
	if err != nil {
		log.Println("schemaFields Get error: ", err)
	}
	for schemaFields.Next() {
		objectConfig = &graphql.Field{
			Type: mapCueStructToGraphQLObject(schemaFields.Value()),
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id := p.Args["id"].(string)
				return resolver.Author(id)
			},
		}
	}

	return objectConfig
}

func mapCueStructToGraphQLObject(cueStruct cue.Value) (graphqlObject graphql.Output) {
    objectName, _ := cueStruct.Label()
	fieldsConfig := graphql.Fields{}
	field, err := cueStruct.Value().Fields()
	if err != nil {
		log.Println("field Get error: ", err)
	}
	for field.Next() {
		fieldsConfig[field.Label()] = mapCueFieldToGraphQLField(field.Value())
	}

	graphqlObject = graphql.NewObject(
		graphql.ObjectConfig{
			Name: objectName,
			Fields: fieldsConfig,
		},
	)

	return
}

func mapCueFieldToGraphQLField(cueType cue.Value) (*graphql.Field) {
	return &graphql.Field{
		Type: mapCueTypeToGraphQLType(cueType),
	}

}

func mapCueListToGraphQLList(cueType cue.Value) (*graphql.List) {
	return graphql.NewList( mapCueTypeToGraphQLType(cueType))
}

func mapCueTypeToGraphQLType(cueType cue.Value) (graphQLType graphql.Output) {
	cueType, _ = cueType.Default()
	fieldName, _ := cueType.Label()
	switch cueType.Kind() {
		case cue.BoolKind:
			log.Println(fieldName, "is of type BoolKind")
			graphQLType = graphql.Boolean
			return

		case cue.IntKind:
			log.Println(fieldName, "is of type IntKind")
			graphQLType = graphql.Int
			return

		case cue.FloatKind:
			log.Println(fieldName, "is of type FloatKind")
			graphQLType = graphql.Float
			return

		case cue.StringKind:
			cueTypeStr, _ := cueType.String()
			// TODO Replace with https://github.com/relvacode/iso8601
			if fieldName == "createdat" {
				parsedDatetime, isoErr := iso8601.ParseString(cueTypeStr)
				if isoErr != nil {
					log.Println(fieldName, "failed iso8601 mapping: ", isoErr)
				} else {
					log.Println(fieldName, "value: ", cueTypeStr, " iso8601 datetime mapping: ", parsedDatetime)
				}
			}

			_, err := time.Parse("2006-03-02T07:20:50.52Z", cueTypeStr)
			if err == nil {
				log.Println(fieldName, "is of type time library DateTime: ", cueTypeStr)
				graphQLType = graphql.DateTime
				return mapCueStructToGraphQLObject(cueType)
			}
			log.Println(fieldName, "is of type StringKind")
			graphQLType = graphql.String
			return

		case cue.ListKind:
			defaultVal,_ := cueType.Default()
			listElems, err := defaultVal.List()
			if err != nil {
				log.Println("Failed retrieving first default value from ListKind cue.field. Field name: ", fieldName, "error: ", err)
			}
			listElems.Next()
			listElemDefault, _ := listElems.Value().Default()

			return graphql.NewList(mapCueTypeToGraphQLType(listElemDefault))

		case cue.StructKind:
			return mapCueStructToGraphQLObject(cueType)

		case cue.NullKind:
			log.Println("Error: No cue schema default value set for field: ", fieldName)

		default:
			log.Println("error: failed attempt to determine type of field: ", fieldName)
	}

	return
}
