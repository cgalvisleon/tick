package store

import (
	"time"

	"github.com/cgalvisleon/et/et"
	"github.com/cgalvisleon/et/jsql"
	"github.com/cgalvisleon/et/reg"
)

// Project wraps the single project row and its tags. A .tick/ directory
// holds exactly one project, so there is never a lookup by id — the store
// always has zero or one row.
type Project struct {
	store *jsql.Model
	tags  *jsql.Model
}

type ProjectInfo struct {
	ID          string
	Code        string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Tags        map[string]string
}

func defineProject(db *jsql.DB) (*Project, error) {
	def := jsql.Def{
		Schema:  schema,
		Name:    "project",
		Version: 1,
		Columns: []jsql.Column{
			{Name: "id", TypeColumn: jsql.COLUMN, TypeData: jsql.KEY},
			{Name: "code", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "name", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "description", TypeColumn: jsql.COLUMN, TypeData: jsql.MEMO},
			{Name: "created_at", TypeColumn: jsql.COLUMN, TypeData: jsql.DATETIME},
			{Name: "updated_at", TypeColumn: jsql.COLUMN, TypeData: jsql.DATETIME},
		},
		PrimaryKeys: []jsql.DefIndex{{Name: "id"}},
	}
	store, err := db.Define(def)
	if err != nil {
		return nil, err
	}
	if err := store.Init(); err != nil {
		return nil, err
	}

	tagsDef := jsql.Def{
		Schema:  schema,
		Name:    "project_tags",
		Version: 1,
		Columns: []jsql.Column{
			{Name: "id", TypeColumn: jsql.COLUMN, TypeData: jsql.KEY},
			{Name: "name", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "value", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
		},
		PrimaryKeys: []jsql.DefIndex{{Name: "id"}},
		Unique:      []jsql.DefIndex{{Name: "name"}},
	}
	tags, err := db.Define(tagsDef)
	if err != nil {
		return nil, err
	}
	if err := tags.Init(); err != nil {
		return nil, err
	}

	return &Project{store: store, tags: tags}, nil
}

/**
* Get: Returns the current project's data, or ok=false when init hasn't set one yet.
* @return ProjectInfo, bool, error
**/
func (s *Project) Get() (ProjectInfo, bool, error) {
	item, err := s.store.Where(jsql.NotNull("id")).One()
	if err != nil {
		return ProjectInfo{}, false, err
	}
	if item.IsEmpty() {
		return ProjectInfo{}, false, nil
	}

	result := ProjectInfo{
		ID:          item.Str("id"),
		Code:        item.Str("code"),
		Name:        item.Str("name"),
		Description: item.Str("description"),
		CreatedAt:   item.Time("created_at"),
		UpdatedAt:   item.Time("updated_at"),
	}

	tags, err := s.Tags()
	if err != nil {
		return ProjectInfo{}, false, err
	}
	result.Tags = tags

	return result, true, nil
}

/**
* Init: Creates the single project row if it doesn't exist yet.
* @return error
**/
func (s *Project) Init() error {
	_, exists, err := s.Get()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	now := time.Now()
	_, err = s.store.Insert(et.Json{
		"id":         reg.UUID(),
		"code":       "",
		"name":       "",
		"description": "",
		"created_at": now,
		"updated_at": now,
	}).One()
	return err
}

/**
* Set: Updates the given fields on the current project row.
* @param fields map[string]string
* @return error
**/
func (s *Project) Set(fields map[string]string) error {
	if len(fields) == 0 {
		return nil
	}

	info, exists, err := s.Get()
	if err != nil {
		return err
	}
	if !exists {
		return jsql.ErrRecordAlreadyExists
	}

	data := et.Json{"id": info.ID, "updated_at": time.Now()}
	for k, v := range fields {
		data[k] = v
	}
	_, err = s.store.Update(data).One()
	return err
}

/**
* Tags: Returns all project tags as a name→value map.
* @return map[string]string, error
**/
func (s *Project) Tags() (map[string]string, error) {
	items, err := s.tags.Where(jsql.NotNull("id")).All()
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for _, item := range items.Result {
		result[item.Str("name")] = item.Str("value")
	}
	return result, nil
}

/**
* SetTag: Creates or updates a project tag.
* @param name string, value string
* @return error
**/
func (s *Project) SetTag(name, value string) error {
	exists, err := s.tags.Where(jsql.Eq("name", name)).Exists()
	if err != nil {
		return err
	}
	if exists {
		_, err = s.tags.Update(et.Json{"value": value}).Where(jsql.Eq("name", name)).One()
		return err
	}
	_, err = s.tags.Insert(et.Json{"id": reg.UUID(), "name": name, "value": value}).One()
	return err
}

/**
* RemoveTag: Deletes a project tag by name.
* @param name string
* @return error
**/
func (s *Project) RemoveTag(name string) error {
	_, err := s.tags.Delete().Where(jsql.Eq("name", name)).One()
	return err
}
