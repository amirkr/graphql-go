package resolver

import (
	"github.com/amirkr/graphql-go/model"
)

func Author(id string) (model.Author, error) {
	author := model.Author{
		ID        : id,
		Firstname : "Edgar Allan",
		Lastname  : "Poe",
		Createdat : "2021-09-22T07:20:50.52Z",
		Object: struct{Obj_id int; Obj_name string}{
			10, "name",
		},
	}
	return author, nil
}