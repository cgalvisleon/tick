package store

import (
	"github.com/cgalvisleon/et/et"
	"github.com/cgalvisleon/et/jsql"
	"github.com/cgalvisleon/et/reg"
)

// Config is a per-project key/value store (user.name, user.email, token, ...),
// backing the `tick config` command. It replaces the old global
// ~/.tick/config file: config is now local to each project's tick.db, so a
// future `login` command can persist an auth token (key "token") the same way
// any other setting is stored.
type Config struct {
	store *jsql.Model
}

func defineConfig(db *jsql.DB) (*Config, error) {
	def := jsql.Def{
		Schema:  schema,
		Name:    "config",
		Version: 1,
		Columns: []jsql.Column{
			{Name: "id", TypeColumn: jsql.COLUMN, TypeData: jsql.KEY},
			{Name: "name", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "value", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
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
	return &Config{store: store}, nil
}

/**
* Get: Returns the value for key, and whether it is set.
* @param name string
* @return string, bool, error
**/
func (s *Config) Get(name string) (string, bool, error) {
	item, err := s.store.Where(jsql.Eq("name", name)).One()
	if err != nil {
		return "", false, err
	}
	if item.IsEmpty() {
		return "", false, nil
	}
	return item.Str("value"), true, nil
}

/**
* Set: Creates or updates a config key.
* @param name string, value string
* @return error
**/
func (s *Config) Set(name, value string) error {
	exists, err := s.store.Where(jsql.Eq("name", name)).Exists()
	if err != nil {
		return err
	}
	if exists {
		_, err = s.store.Update(et.Json{"value": value}).Where(jsql.Eq("name", name)).One()
		return err
	}
	_, err = s.store.Insert(et.Json{"id": reg.UUID(), "name": name, "value": value}).One()
	return err
}
