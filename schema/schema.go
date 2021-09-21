package schema

import (
	"log"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"

	// "github.com/amirkr/graphql-go/model"
	"time"

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
		created_at: string | *"2017-11-11T07:20:50.52Z"
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
			_, err := time.Parse("2006-03-02T07:20:50.52Z", cueTypeStr)
			if err == nil {
				log.Println(fieldName, "is of type DateTime")
				graphQLType = graphql.DateTime
				return
			}
			log.Println(fieldName, "is of type StringKind")
			graphQLType = graphql.String
			return

		case cue.BytesKind:
			log.Println(fieldName, "is of type BytesKind")

		case cue.ListKind:
			log.Println(fieldName, "is of type ListKind")

		case cue.StructKind:
			log.Println(fieldName, "is of type StructKind")

		case cue.NullKind:
			// TODO Handle Error
			log.Println(fieldName, "is of type NullKind")

		default:
			log.Println(fieldName, "is of type default")
	}

	return
}
