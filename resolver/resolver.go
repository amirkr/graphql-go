package resolver

import (
	"github.com/amirkr/graphql-go/model"
)

func Author(id string) (model.Author, error) {
	author := model.Author{
		ID        : id,
		Firstname : "Edgar Allan",
		Lastname  : "Poe",
	}
	return author, nil
}