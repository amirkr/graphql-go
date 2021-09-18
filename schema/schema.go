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
	sysdataSchema := `
	#sysdata: {
		lkwd_id: int | *0
		isDelete: bool| *false
		price: float | *0.0
		name: string | *""
		databaseType: string | *""
		exportType: string | *""
		exportCategory: string | *""
		{
			if sysdata.databaseType == "apartment" {
				exportType:     "Areas"
				exportCategory: "Apartment"
			}
			if sysdata.databaseType == "public_space" {
				exportType:     "Areas"
				exportCategory: "Apartment"
			}
			if sysdata.databaseType == "animation" {
				exportType:     "Avakin"
				exportCategory: "animation"
			}
		}
		...
	}
	sysdata: #sysdata
	`

	cueSchemaStr := sysdataSchema + `
	#author: {
		id: string | *""
		firstname: string  | *""
		lastname: string | *""
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
		objectName := schemaFields.Label()
		// log.Println(schemaFields.Label(), " ", schemaFields.Value())
		fieldsConfig := graphql.Fields{}
		field, err := schemaFields.Value().Fields()
		if err != nil {
			log.Println("field Get error: ", err)
		}
		for field.Next() {
			fieldsConfig[field.Label()] = &graphql.Field{
				Type: mapCueTypeToGraphQLType(field.Value()),
			}
		}
		objectConfig = &graphql.Field{
			Type: graphql.NewObject(
				graphql.ObjectConfig{
					Name: objectName,
					Fields: fieldsConfig,
				},
			),
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

func mapCueTypeToGraphQLType(cueType cue.Value) (graphQLType graphql.Output) {
	graphQLType = nil

	fieldName, _ := cueType.Label()

	if fieldName == "price" {
		_, err := cueType.Float64()
		if err != nil {
			log.Println("float cast error: ", err)
		}
	}

	if _, err := cueType.String(); err == nil {
		log.Println("field ", fieldName, " is of type string")
		// TODO Check if cueType.Label() contains word id then map to graphql.ID
		// graphQLType = graphql.ID
		// TODO IF Attemot to format to date is successful then map to graphql.DateTime
		// graphQLType = graphql.DateTime
		graphQLType = graphql.String
	} else if _, err := cueType.Int64(); err == nil {
		log.Println("field ", fieldName, " is of type int")
		graphQLType = graphql.Int
	} else if _, err := cueType.Float64(); err == nil {
		log.Println("field ", fieldName, " is of type float")
		graphQLType = graphql.Float
	} else if _, err := cueType.Bool(); err == nil {
		log.Println("field ", fieldName, " is of type bool")
		graphQLType = graphql.Boolean
	} else {
		log.Println("field ", fieldName, " is of type unknown")
	}

	return
}
