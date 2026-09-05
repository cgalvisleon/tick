package store

import (
	"fmt"
	"time"

	"github.com/cgalvisleon/et/et"
	"github.com/cgalvisleon/et/jsql"
	"github.com/cgalvisleon/et/reg"
)

const (
	StatusPending   = "pending"
	StatusInProcess = "in_process"
	StatusStop      = "stop"
	StatusAwait     = "await"
	StatusDone      = "done"
)

// pausingStatuses are statuses that stop the real-time clock while active.
var pausingStatuses = map[string]bool{StatusStop: true, StatusAwait: true}

// ValidStatuses is the ordered list of the five statuses tick understands.
var ValidStatuses = []string{StatusPending, StatusInProcess, StatusStop, StatusAwait, StatusDone}

/**
* NormalizeStatus: Maps free-form user input ("in process", "in-process", ...) to
* one of the five canonical status values.
* @param raw string
* @return string, bool
**/
func NormalizeStatus(raw string) (string, bool) {
	norm := ""
	for _, r := range raw {
		if r == ' ' || r == '-' {
			norm += "_"
		} else {
			norm += string(r)
		}
	}
	for _, s := range ValidStatuses {
		if s == norm {
			return s, true
		}
	}
	return "", false
}

type Task struct {
	store   *jsql.Model
	tags    *jsql.Model
	history *jsql.Model
}

type TaskInfo struct {
	ID            string
	Code          string
	Name          string
	Description   string
	Type          string
	Assignee      string
	Status        string
	PlannedStart  string
	PlannedEnd    string
	ActualStart   *time.Time
	ActualEnd     *time.Time
	PausedMinutes int
	ActualMinutes int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Tags          map[string]string
}

type StatusEntry struct {
	Status      string
	Description string
	Percent     int
	CreatedAt   time.Time
}

func defineTask(db *jsql.DB) (*Task, error) {
	def := jsql.Def{
		Schema:  schema,
		Name:    "tasks",
		Version: 1,
		Columns: []jsql.Column{
			{Name: "id", TypeColumn: jsql.COLUMN, TypeData: jsql.KEY},
			{Name: "code", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "name", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "description", TypeColumn: jsql.COLUMN, TypeData: jsql.MEMO},
			{Name: "type", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "assignee", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "status", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT, Default: StatusPending},
			{Name: "planned_start", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "planned_end", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "actual_start", TypeColumn: jsql.COLUMN, TypeData: jsql.DATETIME},
			{Name: "actual_end", TypeColumn: jsql.COLUMN, TypeData: jsql.DATETIME},
			{Name: "paused_minutes", TypeColumn: jsql.COLUMN, TypeData: jsql.INT, Default: 0},
			{Name: "actual_minutes", TypeColumn: jsql.COLUMN, TypeData: jsql.INT, Default: 0},
			{Name: "created_at", TypeColumn: jsql.COLUMN, TypeData: jsql.DATETIME},
			{Name: "updated_at", TypeColumn: jsql.COLUMN, TypeData: jsql.DATETIME},
		},
		PrimaryKeys: []jsql.DefIndex{{Name: "id"}},
		Unique:      []jsql.DefIndex{{Name: "code"}},
		Indexes:     []jsql.DefIndex{{Name: "type"}, {Name: "status"}},
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
		Name:    "task_tags",
		Version: 1,
		Columns: []jsql.Column{
			{Name: "id", TypeColumn: jsql.COLUMN, TypeData: jsql.KEY},
			{Name: "task_id", TypeColumn: jsql.COLUMN, TypeData: jsql.KEY},
			{Name: "name", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "value", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
		},
		PrimaryKeys: []jsql.DefIndex{{Name: "id"}},
		Indexes:     []jsql.DefIndex{{Name: "task_id"}},
	}
	tags, err := db.Define(tagsDef)
	if err != nil {
		return nil, err
	}
	if err := tags.Init(); err != nil {
		return nil, err
	}

	historyDef := jsql.Def{
		Schema:  schema,
		Name:    "task_status_history",
		Version: 1,
		Columns: []jsql.Column{
			{Name: "id", TypeColumn: jsql.COLUMN, TypeData: jsql.KEY},
			{Name: "task_id", TypeColumn: jsql.COLUMN, TypeData: jsql.KEY},
			{Name: "status", TypeColumn: jsql.COLUMN, TypeData: jsql.TEXT},
			{Name: "description", TypeColumn: jsql.COLUMN, TypeData: jsql.MEMO},
			{Name: "percent", TypeColumn: jsql.COLUMN, TypeData: jsql.INT, Default: 0},
			{Name: "created_at", TypeColumn: jsql.COLUMN, TypeData: jsql.DATETIME},
		},
		PrimaryKeys: []jsql.DefIndex{{Name: "id"}},
		Indexes:     []jsql.DefIndex{{Name: "task_id"}},
	}
	history, err := db.Define(historyDef)
	if err != nil {
		return nil, err
	}
	if err := history.Init(); err != nil {
		return nil, err
	}

	return &Task{store: store, tags: tags, history: history}, nil
}

func toInfo(item et.Json) TaskInfo {
	result := TaskInfo{
		ID:            item.Str("id"),
		Code:          item.Str("code"),
		Name:          item.Str("name"),
		Description:   item.Str("description"),
		Type:          item.Str("type"),
		Assignee:      item.Str("assignee"),
		Status:        item.Str("status"),
		PlannedStart:  item.Str("planned_start"),
		PlannedEnd:    item.Str("planned_end"),
		PausedMinutes: item.Int("paused_minutes"),
		ActualMinutes: item.Int("actual_minutes"),
		CreatedAt:     item.Time("created_at"),
		UpdatedAt:     item.Time("updated_at"),
	}
	if item.Exist("actual_start") && item.Str("actual_start") != "" {
		t := item.Time("actual_start")
		result.ActualStart = &t
	}
	if item.Exist("actual_end") && item.Str("actual_end") != "" {
		t := item.Time("actual_end")
		result.ActualEnd = &t
	}
	return result
}

/**
* Find: Looks up a task by id or by code.
* @param id string, code string
* @return TaskInfo, bool, error
**/
func (s *Task) Find(id, code string) (TaskInfo, bool, error) {
	var item et.Item
	var err error
	switch {
	case id != "":
		item, err = s.store.Where(jsql.Eq("id", id)).One()
	case code != "":
		item, err = s.store.Where(jsql.Eq("code", code)).One()
	default:
		return TaskInfo{}, false, fmt.Errorf("id or code is required")
	}
	if err != nil {
		return TaskInfo{}, false, err
	}
	if item.IsEmpty() {
		return TaskInfo{}, false, nil
	}

	info := toInfo(item.Result)
	tags, err := s.Tags(info.ID)
	if err != nil {
		return TaskInfo{}, false, err
	}
	info.Tags = tags
	return info, true, nil
}

/**
* List: Returns every task, most recently created first.
* @return []TaskInfo, error
**/
func (s *Task) List() ([]TaskInfo, error) {
	items, err := s.store.Where(jsql.NotNull("id")).OrderBy("created_at", true).All()
	if err != nil {
		return nil, err
	}
	result := make([]TaskInfo, 0, len(items.Result))
	for _, item := range items.Result {
		result = append(result, toInfo(item))
	}
	return result, nil
}

/**
* Create: Inserts a new task with the given code and fields.
* @param code string, fields map[string]string
* @return TaskInfo, error
**/
func (s *Task) Create(code string, fields map[string]string) (TaskInfo, error) {
	if code == "" {
		return TaskInfo{}, fmt.Errorf("code es requerido")
	}
	if fields["name"] == "" {
		return TaskInfo{}, fmt.Errorf("name es requerido")
	}

	now := time.Now()
	data := et.Json{
		"id":         reg.UUID(),
		"code":       code,
		"status":     StatusPending,
		"created_at": now,
		"updated_at": now,
	}
	for k, v := range fields {
		data[k] = v
	}
	if _, err := s.store.Insert(data).One(); err != nil {
		return TaskInfo{}, err
	}

	info, exists, err := s.Find("", code)
	if err != nil {
		return TaskInfo{}, err
	}
	if !exists {
		return TaskInfo{}, fmt.Errorf("no se pudo crear la tarea")
	}
	return info, nil
}

/**
* Update: Updates the given fields on an existing task.
* @param id string, fields map[string]string
* @return error
**/
func (s *Task) Update(id string, fields map[string]string) error {
	if len(fields) == 0 {
		return nil
	}
	if v, ok := fields["name"]; ok && v == "" {
		return fmt.Errorf("name es requerido")
	}
	if v, ok := fields["code"]; ok && v == "" {
		return fmt.Errorf("code es requerido")
	}
	data := et.Json{"id": id, "updated_at": time.Now()}
	for k, v := range fields {
		data[k] = v
	}
	_, err := s.store.Update(data).One()
	return err
}

/**
* Tags: Returns all tags for a task as a name→value map.
* @param taskID string
* @return map[string]string, error
**/
func (s *Task) Tags(taskID string) (map[string]string, error) {
	items, err := s.tags.Where(jsql.Eq("task_id", taskID)).All()
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
* SetTag: Creates or updates a tag on a task.
* @param taskID string, name string, value string
* @return error
**/
func (s *Task) SetTag(taskID, name, value string) error {
	exists, err := s.tags.Where(jsql.Eq("task_id", taskID)).And(jsql.Eq("name", name)).Exists()
	if err != nil {
		return err
	}
	if exists {
		_, err = s.tags.Update(et.Json{"value": value}).
			Where(jsql.Eq("task_id", taskID)).And(jsql.Eq("name", name)).One()
		return err
	}
	_, err = s.tags.Insert(et.Json{"id": reg.UUID(), "task_id": taskID, "name": name, "value": value}).One()
	return err
}

/**
* RemoveTag: Deletes a tag from a task by name.
* @param taskID string, name string
* @return error
**/
func (s *Task) RemoveTag(taskID, name string) error {
	_, err := s.tags.Delete().Where(jsql.Eq("task_id", taskID)).And(jsql.Eq("name", name)).One()
	return err
}

/**
* History: Returns the status history of a task, oldest first.
* @param taskID string
* @return []StatusEntry, error
**/
func (s *Task) History(taskID string) ([]StatusEntry, error) {
	items, err := s.history.Where(jsql.Eq("task_id", taskID)).OrderBy("created_at", true).All()
	if err != nil {
		return nil, err
	}
	result := make([]StatusEntry, 0, len(items.Result))
	for _, item := range items.Result {
		result = append(result, StatusEntry{
			Status:      item.Str("status"),
			Description: item.Str("description"),
			Percent:     item.Int("percent"),
			CreatedAt:   item.Time("created_at"),
		})
	}
	return result, nil
}

/**
* SetStatus: Appends a status-history entry for the task and updates the task's
* running time-tracking fields (actual_start/actual_end/paused_minutes/actual_minutes)
* according to the pause-aware clock: the clock starts on the first transition into
* in_process, stop/await intervals are subtracted, and it stops on done.
* @param taskID string, status string, description string, percent int
* @return TaskInfo, error
**/
func (s *Task) SetStatus(taskID, status, description string, percent int) (TaskInfo, error) {
	info, exists, err := s.Find(taskID, "")
	if err != nil {
		return TaskInfo{}, err
	}
	if !exists {
		return TaskInfo{}, fmt.Errorf("task not found: %s", taskID)
	}

	now := time.Now()
	previousStatus := info.Status

	_, err = s.history.Insert(et.Json{
		"id":          reg.UUID(),
		"task_id":     taskID,
		"status":      status,
		"description": description,
		"percent":     percent,
		"created_at":  now,
	}).One()
	if err != nil {
		return TaskInfo{}, err
	}

	fields := et.Json{"status": status, "updated_at": now}

	if status == StatusInProcess && info.ActualStart == nil {
		fields["actual_start"] = now
	}

	if pausingStatuses[previousStatus] && !pausingStatuses[status] {
		lastPauseStart, err := s.lastPauseStart(taskID)
		if err == nil && lastPauseStart != nil {
			fields["paused_minutes"] = info.PausedMinutes + int(now.Sub(*lastPauseStart).Minutes())
		}
	}

	if status == StatusDone {
		fields["actual_end"] = now
		if info.ActualStart != nil {
			pausedMinutes := info.PausedMinutes
			if v, ok := fields["paused_minutes"]; ok {
				pausedMinutes = v.(int)
			}
			totalMinutes := int(now.Sub(*info.ActualStart).Minutes()) - pausedMinutes
			if totalMinutes < 0 {
				totalMinutes = 0
			}
			fields["actual_minutes"] = totalMinutes
		}
	}

	fields["id"] = taskID
	if _, err := s.store.Update(fields).One(); err != nil {
		return TaskInfo{}, err
	}

	updated, _, err := s.Find(taskID, "")
	return updated, err
}

// lastPauseStart returns when the task's current stop/await interval began, i.e.
// the created_at of the most recent history entry that transitioned into a
// pausing status.
func (s *Task) lastPauseStart(taskID string) (*time.Time, error) {
	history, err := s.History(taskID)
	if err != nil {
		return nil, err
	}
	for i := len(history) - 1; i >= 0; i-- {
		if pausingStatuses[history[i].Status] {
			t := history[i].CreatedAt
			return &t, nil
		}
		break
	}
	return nil, nil
}

// TypeAverage summarizes done-task durations for one task type.
type TypeAverage struct {
	Type          string
	Count         int
	AvgMinutes    float64
}

/**
* AveragesByType: Computes the average actual duration (minutes) per task type,
* over tasks currently in status done. Computed on read so it never drifts out of
* sync with the underlying history.
* @return []TypeAverage, error
**/
func (s *Task) AveragesByType() ([]TypeAverage, error) {
	items, err := s.store.Where(jsql.Eq("status", StatusDone)).All()
	if err != nil {
		return nil, err
	}

	sums := map[string]int{}
	counts := map[string]int{}
	for _, item := range items.Result {
		tp := item.Str("type")
		if tp == "" {
			tp = "(sin tipo)"
		}
		sums[tp] += item.Int("actual_minutes")
		counts[tp]++
	}

	result := make([]TypeAverage, 0, len(counts))
	for tp, count := range counts {
		result = append(result, TypeAverage{
			Type:       tp,
			Count:      count,
			AvgMinutes: float64(sums[tp]) / float64(count),
		})
	}
	return result, nil
}
