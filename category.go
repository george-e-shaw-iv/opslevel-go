package opslevel

import (
	"fmt"

	"github.com/gosimple/slug"
	"github.com/shurcooL/graphql"
)

type Category struct {
	Id   graphql.ID `json:"id"`
	Name string
}

type CategoryConnection struct {
	Nodes      []Category
	PageInfo   PageInfo
	TotalCount graphql.Int
}

type CategoryCreateInput struct {
	Name string `json:"name"`
}

type CategoryUpdateInput struct {
	Id   graphql.ID `json:"id"`
	Name string     `json:"name"`
}

type CategoryDeleteInput struct {
	Id graphql.ID `json:"id"`
}

func (self *Category) Alias() string {
	return slug.Make(self.Name)
}

func (conn *CategoryConnection) Hydrate(client *Client) error {
	var q struct {
		Account struct {
			Rubric struct {
				Categories CategoryConnection `graphql:"categories(after: $after, first: $first)"`
			}
		}
	}
	v := PayloadVariables{
		"first": client.pageSize,
	}
	q.Account.Rubric.Categories.PageInfo = conn.PageInfo
	for q.Account.Rubric.Categories.PageInfo.HasNextPage {
		v["after"] = q.Account.Rubric.Categories.PageInfo.End
		if err := client.Query(&q, v); err != nil {
			return err
		}
		for _, item := range q.Account.Rubric.Categories.Nodes {
			conn.Nodes = append(conn.Nodes, item)
		}
	}
	return nil
}

//#region Create

func (client *Client) CreateCategory(input CategoryCreateInput) (*Category, error) {
	var m struct {
		Payload struct {
			Category Category
			Errors   []OpsLevelErrors
		} `graphql:"categoryCreate(input: $input)"`
	}
	v := PayloadVariables{
		"input": input,
	}
	if err := client.Mutate(&m, v); err != nil {
		return nil, err
	}
	return &m.Payload.Category, FormatErrors(m.Payload.Errors)
}

//#endregion

//#region Retrieve

func (client *Client) GetCategory(id graphql.ID) (*Category, error) {
	var q struct {
		Account struct {
			Category Category `graphql:"category(id: $id)"`
		}
	}
	v := PayloadVariables{
		"id": id,
	}
	if err := client.Query(&q, v); err != nil {
		return nil, err
	}
	if q.Account.Category.Id == nil {
		return nil, fmt.Errorf("Category with ID '%s' not found!", id)
	}
	return &q.Account.Category, nil
}

func (client *Client) ListCategories() ([]Category, error) {
	var q struct {
		Account struct {
			Rubric struct {
				Categories CategoryConnection
			}
		}
	}
	if err := client.Query(&q, nil); err != nil {
		return q.Account.Rubric.Categories.Nodes, err
	}
	if err := q.Account.Rubric.Categories.Hydrate(client); err != nil {
		return q.Account.Rubric.Categories.Nodes, err
	}
	return q.Account.Rubric.Categories.Nodes, nil
}

//#endregion

//#region Update

func (client *Client) UpdateCategory(input CategoryUpdateInput) (*Category, error) {
	var m struct {
		Payload struct {
			Category Category
			Errors   []OpsLevelErrors
		} `graphql:"categoryUpdate(input: $input)"`
	}
	v := PayloadVariables{
		"input": input,
	}
	if err := client.Mutate(&m, v); err != nil {
		return nil, err
	}
	return &m.Payload.Category, FormatErrors(m.Payload.Errors)
}

//#endregion

//#region Delete

func (client *Client) DeleteCategory(id graphql.ID) error {
	var m struct {
		Payload struct {
			Id     graphql.ID `graphql:"deletedCategoryId"`
			Errors []OpsLevelErrors
		} `graphql:"categoryDelete(input: $input)"`
	}
	v := PayloadVariables{
		"input": CategoryDeleteInput{Id: id},
	}
	if err := client.Mutate(&m, v); err != nil {
		return err
	}
	return FormatErrors(m.Payload.Errors)
}

//#endregion
