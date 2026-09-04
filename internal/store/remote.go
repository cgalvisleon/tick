package store

import (
	"time"

	"github.com/cgalvisleon/et/et"
	"github.com/cgalvisleon/et/jsql"
	"github.com/cgalvisleon/et/reg"
)

type Remote struct {
	store *jsql.Model
}

type RemoteInfo struct {
	Name      string
	Path      string
	CreatedAt time.Time
}

func defineRemote(db *jsql.DB) (*Remote, error) {
	def := jsql.Def{
		Schema:  schema,
		Name:    "remotes",
		Version: 1,
		Columns: []jsql.Column{
			{Name: "id", TypeColumn: jsql.COLUMN, TypeData: jsql.KEY},
			{Name: "name", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "path", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "created_at", TypeColumn: jsql.COLUMN, TypeData: jsql.DATETIME},
		},
		PrimaryKeys: []jsql.DefIndex{{Name: "id"}},
		Unique:      []jsql.DefIndex{{Name: "name"}},
	}
	store, err := db.Define(def)
	if err != nil {
		return nil, err
	}
	if err := store.Init(); err != nil {
		return nil, err
	}
	return &Remote{store: store}, nil
}

/**
* List: Returns every configured remote.
* @return []RemoteInfo, error
**/
func (s *Remote) List() ([]RemoteInfo, error) {
	items, err := s.store.Where(jsql.NotNull("id")).All()
	if err != nil {
		return nil, err
	}
	result := make([]RemoteInfo, 0, len(items.Result))
	for _, item := range items.Result {
		result = append(result, RemoteInfo{
			Name:      item.Str("name"),
			Path:      item.Str("path"),
			CreatedAt: item.Time("created_at"),
		})
	}
	return result, nil
}

/**
* Get: Looks up a remote by name.
* @param name string
* @return RemoteInfo, bool, error
**/
func (s *Remote) Get(name string) (RemoteInfo, bool, error) {
	item, err := s.store.Where(jsql.Eq("name", name)).One()
	if err != nil {
		return RemoteInfo{}, false, err
	}
	if item.IsEmpty() {
		return RemoteInfo{}, false, nil
	}
	return RemoteInfo{
		Name:      item.Str("name"),
		Path:      item.Str("path"),
		CreatedAt: item.Time("created_at"),
	}, true, nil
}

/**
* Add: Creates or updates a remote's path.
* @param name string, path string
* @return error
**/
func (s *Remote) Add(name, path string) error {
	_, exists, err := s.Get(name)
	if err != nil {
		return err
	}
	if exists {
		_, err = s.store.Update(et.Json{"path": path}).Where(jsql.Eq("name", name)).One()
		return err
	}
	_, err = s.store.Insert(et.Json{
		"id":         reg.UUID(),
		"name":       name,
		"path":       path,
		"created_at": time.Now(),
	}).One()
	return err
}

/**
* Remove: Deletes a remote by name.
* @param name string
* @return error
**/
func (s *Remote) Remove(name string) error {
	_, err := s.store.Delete().Where(jsql.Eq("name", name)).One()
	return err
}
