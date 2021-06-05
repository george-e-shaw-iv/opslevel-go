package opslevel

import (
	"fmt"

	"github.com/shurcooL/graphql"
)

type Language struct {
	Name  string
	Usage float32
}

type RepositoryId struct {
	Id graphql.ID
}

type Repository struct {
	// https://pkg.go.dev/github.com/relvacode/iso8601
	//CreatedOn ISO8601DateTime `json:",omitempty"`
	DefaultBranch string `json:",omitempty"`
	Description   string `json:",omitempty"`
	Forked        bool   `json:",omitempty"`
	HtmlUrl       string
	RepositoryId
	Languages []Language
	// LastOwnerChangedAt ISO8601DateTime
	Name         string `json:",omitempty"`
	Organization string
	Owner        Team `json:",omitempty"`
	Private      bool `json:",omitempty"`
	RepoKey      string
	Services     RepositoryServiceConnection
	Tags         RepositoryTagConnection
	Tier         Tier
	Type         string
	Url          string `json:",omitempty"`
	Visible      bool   `json:",omitempty"`
}

type RepositoryConnection struct {
	HiddenCount       int
	Nodes             []Repository
	OrganizationCount int
	OwnedCount        int
	PageInfo          PageInfo
	TotalCount        int
	VisibleCount      int
}

type RepositoryServiceConnection struct {
	Nodes      []ServiceId
	PageInfo   PageInfo
	TotalCount graphql.Int
}

type RepositoryTagConnection struct {
	Nodes      []Tag
	PageInfo   PageInfo
	TotalCount graphql.Int
}

type ServiceRepository struct {
	BaseDirectory string
	DisplayName   string
	Id            graphql.ID
	Repository    RepositoryId
	Service       ServiceId
}

type ServiceRepositoryCreateInput struct {
	Service       IdentifierInput `json:"service"`
	Repository    IdentifierInput `json:"repository"`
	BaseDirectory string          `json:"baseDirectory"`
	DisplayName   string          `json:"displayName,omitempty"`
}

func (r *Repository) Hydrate(client *Client) error {
	if err := r.Services.Hydrate(r.Id, client); err != nil {
		return err
	}
	if err := r.Tags.Hydrate(r.Id, client); err != nil {
		return err
	}
	return nil
}

//#region Create

func (client *Client) ConnectServiceRepository(service *ServiceId, repository *Repository) (*ServiceRepository, error) {
	input := ServiceRepositoryCreateInput{
		Service:       IdentifierInput{Id: service.Id},
		Repository:    IdentifierInput{Id: repository.Id},
		BaseDirectory: "/",
		DisplayName:   fmt.Sprintf("%s/%s", repository.Name, repository.Organization),
	}
	return client.CreateServiceRepository(input)
}

func (client *Client) CreateServiceRepository(input ServiceRepositoryCreateInput) (*ServiceRepository, error) {
	var m struct {
		Payload struct {
			ServiceRepository ServiceRepository
			Errors            []OpsLevelErrors
		} `graphql:"serviceRepositoryCreate(input: $input)"`
	}
	v := PayloadVariables{
		"input": input,
	}
	if err := client.Mutate(&m, v); err != nil {
		return nil, err
	}
	return &m.Payload.ServiceRepository, FormatErrors(m.Payload.Errors)
}

//#endregion

//#region Retrieve

func (client *Client) GetRepositoryWithAlias(alias string) (*Repository, error) {
	var q struct {
		Account struct {
			Repository Repository `graphql:"repository(alias: $repo)"`
		}
	}
	v := PayloadVariables{
		"repo": graphql.String(alias),
	}
	if err := client.Query(&q, v); err != nil {
		return nil, err
	}
	if err := q.Account.Repository.Hydrate(client); err != nil {
		return &q.Account.Repository, err
	}
	return &q.Account.Repository, nil
}

func (client *Client) GetRepository(id graphql.ID) (*Repository, error) {
	var q struct {
		Account struct {
			Repository Repository `graphql:"repository(id: $repo)"`
		}
	}
	v := PayloadVariables{
		"repo": id,
	}
	if err := client.Query(&q, v); err != nil {
		return nil, err
	}
	if err := q.Account.Repository.Hydrate(client); err != nil {
		return &q.Account.Repository, err
	}
	return &q.Account.Repository, nil
}

func (conn *RepositoryConnection) Hydrate(client *Client) error {
	var q struct {
		Account struct {
			Repositories RepositoryConnection `graphql:"repositories(after: $after, first: $first)"`
		}
	}
	v := PayloadVariables{
		"first": client.pageSize,
	}
	q.Account.Repositories.PageInfo = conn.PageInfo
	for q.Account.Repositories.PageInfo.HasNextPage {
		v["after"] = q.Account.Repositories.PageInfo.End
		if err := client.Query(&q, v); err != nil {
			return err
		}
		for _, item := range q.Account.Repositories.Nodes {
			if err := (&item).Hydrate(client); err != nil {
				return err
			}
			conn.Nodes = append(conn.Nodes, item)
		}
	}
	return nil
}

func (conn *RepositoryServiceConnection) Hydrate(id graphql.ID, client *Client) error {
	var q struct {
		Account struct {
			Repository struct {
				Services RepositoryServiceConnection `graphql:"services(after: $after, first: $first)"`
			} `graphql:"repository(id: $id)"`
		}
	}
	v := PayloadVariables{
		"id":    id,
		"first": client.pageSize,
	}
	q.Account.Repository.Services.PageInfo = conn.PageInfo
	for q.Account.Repository.Services.PageInfo.HasNextPage {
		v["after"] = q.Account.Repository.Services.PageInfo.End
		if err := client.Query(&q, v); err != nil {
			return err
		}
		for _, item := range q.Account.Repository.Services.Nodes {
			conn.Nodes = append(conn.Nodes, item)
		}
	}
	return nil
}

func (conn *RepositoryTagConnection) Hydrate(id graphql.ID, client *Client) error {
	var q struct {
		Account struct {
			Repository struct {
				Tags RepositoryTagConnection `graphql:"tags(after: $after, first: $first)"`
			} `graphql:"repository(id: $id)"`
		}
	}
	v := PayloadVariables{
		"id":    id,
		"first": client.pageSize,
	}
	q.Account.Repository.Tags.PageInfo = conn.PageInfo
	for q.Account.Repository.Tags.PageInfo.HasNextPage {
		v["after"] = q.Account.Repository.Tags.PageInfo.End
		if err := client.Query(&q, v); err != nil {
			return err
		}
		for _, item := range q.Account.Repository.Tags.Nodes {
			conn.Nodes = append(conn.Nodes, item)
		}
	}
	return nil
}

func (client *Client) ListRepositories() ([]Repository, error) {
	var q struct {
		Account struct {
			Repositories RepositoryConnection `graphql:"repositories(after: $after, first: $first)"`
		}
	}
	v := PayloadVariables{
		"after": graphql.String(""),
		"first": client.pageSize,
	}
	if err := client.Query(&q, v); err != nil {
		return q.Account.Repositories.Nodes, err
	}
	if err := q.Account.Repositories.Hydrate(client); err != nil {
		return q.Account.Repositories.Nodes, err
	}
	return q.Account.Repositories.Nodes, nil
}
