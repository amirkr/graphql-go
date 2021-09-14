package schema

import (
	"fmt"
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
		databaseType: string | "apartment" | "stg" | "animation"
		...
	}
	sysdata: #sysdata
	`
	// log.Println("sysdataSchema: ", sysdataSchema)

	cueSchemaStr := sysdataSchema + `
	#author: {
		id: string
		firstname: string
		lastname: string
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
		// log.Println("schemaFields IsDefinition? ", schemaFields.IsDefinition())
		// log.Println(schemaFields.Label(), " ", schemaFields.Value())
		fieldsConfig := graphql.Fields{}
		if schemaFields.Label() == "author" {
			field, err := schemaFields.Value().Fields()
			if err != nil {
				log.Println("field Get error: ", err)
			}
			for field.Next() {
				// log.Println("author ", field.Label(), " ", field.Value())
				fieldsConfig[field.Label()] = &graphql.Field{
					Type: mapCueTypeToGraphQLType(field.Value()),
				}
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
		log.Printf("NEW v4 objectConfigType: %T", objectConfig)
	}
	return objectConfig
}

func mapCueTypeToGraphQLType(cueType cue.Value) graphql.Output {
	switch fmt.Sprintf("%t", cueType) {
		case "int":
			log.Println()
			return graphql.Int
		case "string":
			return graphql.String
		default:
			return nil
	}
}