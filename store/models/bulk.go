package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartcontractkit/chainlink/utils"
)

// BulkTaskStatus indicates what a bulk task is doing.
type BulkTaskStatus string

const (
	// BulkTaskStatusInProgress is the default state of any run status.
	BulkTaskStatusInProgress = BulkTaskStatus("")
	// BulkTaskStatusErrored means a bulk task stopped because it encountered an error.
	BulkTaskStatusErrored = BulkTaskStatus("errored")
	// BulkTaskStatusCompleted means a bulk task finished.
	BulkTaskStatusCompleted = BulkTaskStatus("completed")
)

func (t BulkTaskStatus) Value() (driver.Value, error) {
	return string(t), nil
}

func (t *BulkTaskStatus) Scan(value interface{}) error {
	temp, ok := value.([]uint8)
	if !ok {
		return fmt.Errorf("Unable to convert %v of %T to BulkTaskStatus", value, value)
	}

	*t = BulkTaskStatus(temp)
	return nil
}

// BulkDeleteRunRequest describes the query for deletion of runs
type BulkDeleteRunRequest struct {
	gorm.Model
	Status        RunStatusCollection `json:"status" gorm:"type:text"`
	UpdatedBefore time.Time           `json:"updatedBefore"`
}

// BulkDeleteRunTask represents a task that is working to delete runs with a query
type BulkDeleteRunTask struct {
	ID           string               `json:"id" gorm:"primary_key"`
	Query        BulkDeleteRunRequest `json:"query"`
	QueryID      uint                 `json:"-"`
	Status       BulkTaskStatus       `json:"status"`
	ErrorMessage string               `json:"error" gorm:"type:varchar(255)"`
	CreatedAt    time.Time
}

// NewBulkDeleteRunTask returns a task from a request to make a task
func NewBulkDeleteRunTask(request BulkDeleteRunRequest) (*BulkDeleteRunTask, error) {
	for _, status := range request.Status {
		if status != RunStatusCompleted && status != RunStatusErrored {
			return nil, fmt.Errorf("cannot delete Runs with status %s", status)
		}
	}

	return &BulkDeleteRunTask{
		ID:    utils.NewBytes32ID(),
		Query: request,
	}, nil
}

// GetID returns the ID of this structure for jsonapi serialization.
func (t BulkDeleteRunTask) GetID() string {
	return t.ID
}

// GetName returns the pluralized "type" of this structure for jsonapi serialization.
func (t BulkDeleteRunTask) GetName() string {
	return "bulk_delete_runs_tasks"
}

// SetID is used to set the ID of this structure when deserializing from jsonapi documents.
func (t *BulkDeleteRunTask) SetID(value string) error {
	t.ID = value
	return nil
}

type RunStatusCollection []RunStatus

func (r RunStatusCollection) ToStrings() []string {
	// Unable to convert copy-free without unsafe:
	// https://stackoverflow.com/a/48554123/639773
	converted := make([]string, len(r))
	for i, e := range r {
		converted[i] = string(e)
	}
	return converted
}

func (r RunStatusCollection) Value() (driver.Value, error) {
	return strings.Join(r.ToStrings(), ","), nil
}

func (r *RunStatusCollection) Scan(value interface{}) error {
	temp, ok := value.([]uint8)
	if !ok {
		return fmt.Errorf("Unable to convert %v of %T to RunStatusCollection", value, value)
	}

	arr := strings.Split(string(temp), ",")
	collection := make(RunStatusCollection, len(arr))
	for i, r := range arr {
		collection[i] = RunStatus(r)
	}
	*r = collection
	return nil
}
