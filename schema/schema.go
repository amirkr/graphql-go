package schema

import (
	"log"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"

	// "github.com/amirkr/graphql-go/model"

	"github.com/amirkr/graphql-go/resolver"
	"github.com/graphql-go/graphql"
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
			// No dynamic resolver for now as it would have to return results as an UnstructuredJSON
			// and UnstructuredJSON is currently being re-worked not to use gabs json library.
			// https://github.com/Jeffail/gabs library unfortunately doesn't keep order of json attributes
			// matching order of struct fields
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id := p.Args["id"].(string)
				return resolver.Author(id)
			},
		}
	}

	return objectConfig
}

func mapCueStructToGraphQLObject(cueStruct cue.Value) (graphqlObject graphql.Output) {
    objectName, objNameErr := cueStruct.Label()
	if objNameErr != true {
		log.Println("Failed cue.StructKind field name: ", objNameErr)
	}
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

func mapCueListToGraphQLList(cueType cue.Value) (graphqlList *graphql.List) {
	fieldName, objNameErr := cueType.Label()
	if objNameErr != true {
		log.Println("Failed retrieving field name, error ", objNameErr)
	}
	defaultVal, defaultValErr := cueType.Default()
	if defaultValErr != true {
		log.Println("Failed retrieving default value for field: ", fieldName, "error: ", objNameErr)
	}
	listElems, err := defaultVal.List()
	if err != nil {
		log.Println("Failed retrieving first default value from ListKind cue.field. Field name: ", fieldName, "error: ", err)
	}
	listElems.Next()
	listElemDefault, _ := listElems.Value().Default()

	graphqlList = graphql.NewList(mapCueTypeToGraphQLType(listElemDefault))
	return
}

func mapCueTypeToGraphQLType(cueType cue.Value) (graphQLType graphql.Output) {
	fieldName, objNameErr := cueType.Label()
	if objNameErr != true {
		log.Println("Failed retrieving field name, error ", objNameErr)
	}
	cueType, defaultValErr := cueType.Default()
	if defaultValErr != true {
		log.Println("Failed retrieving default value for field: ", fieldName, "error: ", objNameErr)
	}
	switch cueType.Kind() {
		case cue.BoolKind:
			graphQLType = graphql.Boolean
			return

		case cue.IntKind:
			graphQLType = graphql.Int
			return

		case cue.FloatKind:
			graphQLType = graphql.Float
			return

		case cue.StringKind:
			// Not using graphql.DateTime for datetime strings because
			// graphql.DateTime uses RFC 3339 and mongomanager-v2 now uses iso8601
			graphQLType = graphql.String
			return

		case cue.ListKind:
			graphQLType = mapCueListToGraphQLList(cueType)
			return

		case cue.StructKind:
			graphQLType = mapCueStructToGraphQLObject(cueType)
			return

		case cue.NullKind:
			log.Println("Error: No cue schema default value set for field: ", fieldName)

		default:
			log.Println("error: failed attempt to determine type of field: ", fieldName)
	}

	return
}
