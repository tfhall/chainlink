package orm

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/smartcontractkit/chainlink/store/models"
	"github.com/smartcontractkit/chainlink/utils"
	"go.uber.org/multierr"
)

var (
	// ErrorNotFound is returned when finding a single value fails.
	ErrorNotFound = gorm.ErrRecordNotFound
)

// ORM contains the database object used by Chainlink.
type ORM struct {
	DB *gorm.DB
}

// NewORM initializes a new database file at the configured path.
func NewORM(path string) (*ORM, error) {
	db, err := initializeDatabase(path)
	if err != nil {
		return nil, fmt.Errorf("unable to init DB: %+v", err)
	}
	return &ORM{DB: db}, nil
}

func initializeDatabase(path string) (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("unable to open gorm DB: %+v", err)
	}
	return db, nil
}

// TODO: Overkill? Could remove with .Error.
func multify(db *gorm.DB) error {
	var merr error
	for _, e := range db.GetErrors() {
		if e == gorm.ErrRecordNotFound {
			merr = multierr.Append(merr, ErrorNotFound)
		} else {
			merr = multierr.Append(merr, e)
		}
	}
	return merr
}

func multifyWithoutRecordNotFound(db *gorm.DB) error {
	var merr error
	for _, e := range db.GetErrors() {
		if e != gorm.ErrRecordNotFound {
			merr = multierr.Append(merr, e)
		}
	}
	return merr
}

func (orm *ORM) Close() error {
	return orm.DB.Close()
}

// Where fetches multiple objects with "Find".
func (orm *ORM) Where(field string, value interface{}, instance interface{}) error {
	return multify(orm.DB.Where(fmt.Sprintf("%v = ?", field), value).Find(instance))
}

// FindBridge looks up a Bridge by its Name.
func (orm *ORM) FindBridge(name string) (models.BridgeType, error) {
	var bt models.BridgeType
	return bt, multify(orm.DB.Set("gorm:auto_preload", true).First(&bt, "name = ?", name))
}

// PendingBridgeType returns the bridge type of the current pending task,
// or error if not pending bridge.
func (orm *ORM) PendingBridgeType(jr models.JobRun) (models.BridgeType, error) {
	nextTask := jr.NextTaskRun()
	if nextTask == nil {
		return models.BridgeType{}, errors.New("Cannot find the pending bridge type of a job run with no unfinished tasks")
	}
	return orm.FindBridge(nextTask.TaskSpec.Type.String())
}

// FindJob looks up a Job by its ID.
func (orm *ORM) FindJob(id string) (models.JobSpec, error) {
	var job models.JobSpec
	return job, multify(orm.DB.Set("gorm:auto_preload", true).First(&job, "id = ?", id))
}

// FindInitiator returns the single initiator defined by the passed ID.
func (orm *ORM) FindInitiator(ID string) (models.Initiator, error) {
	initr := models.Initiator{}
	return initr, multify(orm.DB.Set("gorm:auto_preload", true).First(&initr, "id = ?", ID))
}

// FindJobRun looks up a JobRun by its ID.
func (orm *ORM) FindJobRun(id string) (models.JobRun, error) {
	var jr models.JobRun
	err := multify(orm.DB.Set("gorm:auto_preload", true).First(&jr, "id = ?", id))
	if err != nil {
		return jr, err
	}

	var rr models.RunResult
	err = multifyWithoutRecordNotFound(orm.DB.First(&rr, "job_run_id = ? AND (task_run_id IS NULL OR task_run_id = ?)", id, ""))
	jr.Result = rr
	return jr, err
}

// SaveJobRun updates UpdatedAt for a JobRun and saves it
func (orm *ORM) SaveJobRun(run *models.JobRun) error {
	return multify(orm.DB.Save(run))
}

// FindServiceAgreement looks up a ServiceAgreement by its ID.
func (orm *ORM) FindServiceAgreement(id string) (models.ServiceAgreement, error) {
	var sa models.ServiceAgreement
	return sa, multify(orm.DB.Set("gorm:auto_preload", true).First(&sa, "id = ?", id))
}

// Jobs fetches all jobs.
func (orm *ORM) Jobs(cb func(models.JobSpec) bool) error {
	db := orm.DB
	offset := 0
	limit := 1000
	for {
		jobs := []models.JobSpec{}
		err := multify(db.Limit(limit).Offset(offset).Find(&jobs))
		if err != nil {
			return err
		}
		for _, j := range jobs {
			if !cb(j) {
				return nil
			}
		}

		if len(jobs) < limit {
			return nil
		}

		offset += limit
	}
}

// JobRunsFor fetches all JobRuns with a given Job ID,
// sorted by their created at time.
func (orm *ORM) JobRunsFor(jobSpecID string) ([]models.JobRun, error) {
	runs := []models.JobRun{}
	err := multify(orm.DB.Set("gorm:auto_preload", true).Where("job_spec_id = ?", jobSpecID).Order("created_at desc").Find(&runs))
	return runs, err
}

// JobRunsCountFor returns the current number of runs for the job
func (orm *ORM) JobRunsCountFor(jobSpecID string) (int, error) {
	var count int
	err := multify(orm.DB.Model(&models.JobRun{}).Where("job_spec_id = ?", jobSpecID).Count(&count))
	return count, err
}

// Sessions returns all sessions limited by the parameters.
func (orm *ORM) Sessions(offset, limit int) ([]models.Session, error) {
	var sessions []models.Session
	err := multify(orm.DB.Set("gorm:auto_preload", true).Limit(limit).Offset(offset).Find(&sessions))
	return sessions, err
}

// SaveJob saves a job to the database and adds IDs to associated tables.
func (orm *ORM) SaveJob(job *models.JobSpec) error {
	for i := range job.Initiators {
		if job.Initiators[i].ID == "" {
			job.Initiators[i].ID = utils.NewBytes32ID()
		}
		job.Initiators[i].JobSpecID = job.ID
	}
	return multify(orm.DB.Save(job))
}

// SaveServiceAgreement saves a service agreement and it's associations to the
// database.
func (orm *ORM) SaveServiceAgreement(sa *models.ServiceAgreement) error {
	merr := multify(orm.DB.Save(sa))
	return merr
}

// JobRunsWithStatus returns the JobRuns which have the passed statuses.
func (orm *ORM) JobRunsWithStatus(statuses ...models.RunStatus) ([]models.JobRun, error) {
	runs := []models.JobRun{}
	merr := multify(orm.DB.Set("gorm:auto_preload", true).Where("status IN (?)", statuses).Find(&runs))
	return runs, merr
}

// AnyJobWithType returns true if there is at least one job associated with
// the type name specified and false otherwise
func (orm *ORM) AnyJobWithType(taskTypeName string) (bool, error) {
	db := orm.DB
	var taskSpec models.TaskSpec
	rval := db.Where("type = ?", taskTypeName).First(&taskSpec)
	found := !rval.RecordNotFound()
	return found, multifyWithoutRecordNotFound(rval)
}

// CreateTx saves the properties of an Ethereum transaction to the database.
func (orm *ORM) CreateTx(
	from common.Address,
	nonce uint64,
	to common.Address,
	data []byte,
	value *big.Int,
	gasLimit uint64,
) (*models.Tx, error) {
	tx := models.Tx{
		From:     from,
		To:       to,
		Nonce:    nonce,
		Data:     data,
		Value:    models.NewBig(value),
		GasLimit: gasLimit,
	}
	return &tx, multify(orm.DB.Save(&tx))
}

// ConfirmTx updates the database for the given transaction to
// show that the transaction has been confirmed on the blockchain.
func (orm *ORM) ConfirmTx(tx *models.Tx, txat *models.TxAttempt) error {
	txat.Confirmed = true
	tx.AssignTxAttempt(txat)
	return multify(orm.DB.Save(tx).Save(txat))
}

// FindTx returns the specific transaction for the passed ID.
func (orm *ORM) FindTx(ID uint64) (*models.Tx, error) {
	tx := &models.Tx{}
	err := multify(orm.DB.Set("gorm:auto_preload", true).First(tx, "id = ?", ID))
	return tx, err
}

// FindTxAttempt returns the specific transaction attempt with the hash.
func (orm *ORM) FindTxAttempt(hash common.Hash) (*models.TxAttempt, error) {
	txat := &models.TxAttempt{}
	err := multify(orm.DB.Set("gorm:auto_preload", true).First(txat, "hash = ?", hash))
	return txat, err
}

// TxAttemptsFor returns the Transaction Attempts (TxAttempt) for a
// given Transaction ID (TxID).
func (orm *ORM) TxAttemptsFor(id uint64) ([]models.TxAttempt, error) {
	attempts := []models.TxAttempt{}
	err := orm.Where("tx_id", id, &attempts)
	return attempts, err
}

// AddTxAttempt creates a new transaction attempt and stores it
// in the database.
func (orm *ORM) AddTxAttempt(
	tx *models.Tx,
	etx *types.Transaction,
	blkNum uint64,
) (*models.TxAttempt, error) {
	hex, err := utils.EncodeTxToHex(etx)
	if err != nil {
		return nil, err
	}
	attempt := &models.TxAttempt{
		Hash:     etx.Hash(),
		GasPrice: models.NewBig(etx.GasPrice()),
		Hex:      hex,
		TxID:     tx.ID,
		SentAt:   blkNum,
	}
	if !tx.Confirmed {
		tx.AssignTxAttempt(attempt)
	}
	err = multify(orm.DB.Save(tx).Save(attempt))
	return attempt, err
}

// GetLastNonce retrieves the last known nonce in the database for an account
func (orm *ORM) GetLastNonce(address common.Address) (uint64, error) {
	var transaction models.Tx
	rval := orm.DB.Order("nonce desc").Where("\"from\" = ?", address).First(&transaction)
	err := multifyWithoutRecordNotFound(rval)
	return transaction.Nonce, err
}

// MarkRan will set Ran to true for a given initiator
func (orm *ORM) MarkRan(i *models.Initiator) error {
	return multify(orm.DB.Model(i).Update("ran", true))
}

// FindUser will return the one API user, or an error.
func (orm *ORM) FindUser() (models.User, error) {
	user := models.User{}
	err := multify(orm.DB.Set("gorm:auto_preload", true).Order("created_at desc").First(&user))
	return user, err
}

// AuthorizedUserWithSession will return the one API user if the Session ID exists
// and hasn't expired, and update session's LastUsed field.
func (orm *ORM) AuthorizedUserWithSession(sessionID string, sessionDuration time.Duration) (models.User, error) {
	if len(sessionID) == 0 {
		return models.User{}, errors.New("Session ID cannot be empty")
	}

	var session models.Session
	err := multify(orm.DB.First(&session, "id = ?", sessionID))
	if err != nil {
		return models.User{}, err
	}
	now := time.Now()
	if session.LastUsed.Time.Add(sessionDuration).Before(now) {
		return models.User{}, errors.New("Session has expired")
	}
	session.LastUsed = models.Time{Time: now}
	if err := multify(orm.DB.Save(&session)); err != nil {
		return models.User{}, err
	}
	return orm.FindUser()
}

// DeleteUser will delete the API User in the db.
func (orm *ORM) DeleteUser() (models.User, error) {
	user, err := orm.FindUser()
	if err != nil {
		return user, err
	}

	tx := orm.DB.Begin()
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		return user, err
	}

	if err := tx.Delete(models.Session{}).Error; err != nil {
		tx.Rollback()
		return user, err
	}

	return user, tx.Commit().Error
}

// DeleteUserSession will erase the session ID for the sole API User.
func (orm *ORM) DeleteUserSession(sessionID string) error {
	return orm.DB.Where("id = ?", sessionID).Delete(models.Session{}).Error
}

// DeleteBridgeType removes the bridge type with passed name.
func (orm *ORM) DeleteBridgeType(name models.TaskType) error {
	return orm.DB.Delete(&models.BridgeType{}, "name = ?", name).Error
}

// DeleteJobRun deletes the job run and corresponding task runs.
func (orm *ORM) DeleteJobRun(ID string) error {
	tx := orm.DB.Begin()
	if err := tx.Where("id = ?", ID).Delete(models.JobRun{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("job_run_id = ?", ID).Delete(models.TaskRun{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// CreateSession will check the password in the SessionRequest against
// the hashed API User password in the db.
func (orm *ORM) CreateSession(sr models.SessionRequest) (string, error) {
	user, err := orm.FindUser()
	if err != nil {
		return "", err
	}

	if !constantTimeEmailCompare(sr.Email, user.Email) {
		return "", errors.New("Invalid email")
	}

	if utils.CheckPasswordHash(sr.Password, user.HashedPassword) {
		session := models.NewSession()
		return session.ID, orm.DB.Save(&session).Error
	}
	return "", errors.New("Invalid password")
}

const constantTimeEmailLength = 256

func constantTimeEmailCompare(left, right string) bool {
	length := utils.MaxInt(constantTimeEmailLength, len(left), len(right))
	leftBytes := make([]byte, length)
	rightBytes := make([]byte, length)
	copy(leftBytes, left)
	copy(rightBytes, right)
	return subtle.ConstantTimeCompare(leftBytes, rightBytes) == 1
}

// ClearSessions removes all sessions.
func (orm *ORM) ClearSessions() error {
	return orm.DB.Delete(models.Session{}).Error
}

// ClearNonCurrentSessions removes all sessions but the id passed in.
func (orm *ORM) ClearNonCurrentSessions(sessionID string) error {
	return orm.DB.Where("id <> ?", sessionID).Delete(models.Session{}).Error
}

// SortType defines the different sort orders available.
type SortType int

const (
	// Ascending is the sort order going up, i.e. 1,2,3.
	Ascending SortType = iota
	// Descending is the sort order going down, i.e. 3,2,1.
	Descending
)

func (s SortType) String() string {
	orderStr := "asc"
	if s == Descending {
		orderStr = "desc"
	}
	return orderStr
}

// JobsSorted returns many JobSpecs sorted by CreatedAt from the store adhering
// to the passed parameters.
func (orm *ORM) JobsSorted(order SortType, offset int, limit int) ([]models.JobSpec, int, error) {
	var count int
	err := orm.DB.Model(&models.JobSpec{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	var jobs []models.JobSpec
	rval := orm.DB.
		Set("gorm:auto_preload", true).
		Order(fmt.Sprintf("created_at %s", order.String())).
		Limit(limit).
		Offset(offset).
		Find(&jobs)
	return jobs, count, rval.Error
}

// TxFrom returns all transactions from a particular address.
func (orm *ORM) TxFrom(from common.Address) ([]models.Tx, error) {
	txs := []models.Tx{}
	return txs, multify(orm.DB.Set("gorm:auto_preload", true).Find(&txs, "\"from\" = ?", from))
}

// Transactions returns all transactions limited by passed parameters.
func (orm *ORM) Transactions(offset, limit int) ([]models.Tx, error) {
	var txs []models.Tx
	err := orm.DB.
		Set("gorm:auto_preload", true).
		Order("id desc").Limit(limit).Offset(offset).
		Find(&txs).Error
	return txs, err
}

// TxAttempts returns the last tx attempts sorted by sent at descending.
func (orm *ORM) TxAttempts(offset, limit int) ([]models.TxAttempt, int, error) {
	var count int
	err := orm.DB.Model(&models.TxAttempt{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	var attempts []models.TxAttempt
	err = orm.DB.
		Set("gorm:auto_preload", true).
		Order("sent_at desc").Limit(limit).Offset(offset).
		Find(&attempts).Error
	return attempts, count, err
}

// JobRunsCount returns the total number of job runs
func (orm *ORM) JobRunsCount() (int, error) {
	var count int
	return count, orm.DB.Model(&models.JobRun{}).Count(&count).Error
}

// JobRunsSorted returns job runs ordered and filtered by the passed params.
func (orm *ORM) JobRunsSorted(order SortType, offset int, limit int) ([]models.JobRun, int, error) {
	count, err := orm.JobRunsCount()
	if err != nil {
		return nil, 0, err
	}

	var runs []models.JobRun
	rval := orm.DB.
		Set("gorm:auto_preload", true).
		Order(fmt.Sprintf("created_at %s", order.String())).
		Limit(limit).Offset(offset).Find(&runs)
	return runs, count, multifyWithoutRecordNotFound(rval)
}

// JobRunsSortedFor returns job runs for a specific job spec ordered and
// filtered by the passed params.
func (orm *ORM) JobRunsSortedFor(id string, order SortType, offset int, limit int) ([]models.JobRun, int, error) {
	count, err := orm.JobRunsCountFor(id)
	if err != nil {
		return nil, 0, err
	}

	var runs []models.JobRun
	rval := orm.DB.
		Set("gorm:auto_preload", true).
		Order(fmt.Sprintf("created_at %s", order.String())).
		Limit(limit).Offset(offset).Find(&runs)
	return runs, count, multifyWithoutRecordNotFound(rval)
}

// BridgeTypes returns bridge types ordered by name filtered limited by the
// passed params.
func (orm *ORM) BridgeTypes(offset int, limit int) ([]models.BridgeType, int, error) {
	db := orm.DB
	var count int
	err := db.Model(&models.BridgeType{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	var bridges []models.BridgeType
	err = db.Order("name asc").Limit(limit).Offset(offset).Find(&bridges).Error
	return bridges, count, err
}

// SaveUser saves the user.
func (orm *ORM) SaveUser(user *models.User) error {
	return orm.DB.Save(user).Error
}

// SaveSession saves the session.
func (orm *ORM) SaveSession(session *models.Session) error {
	return orm.DB.Save(session).Error
}

// SaveBridgeType saves the bridge type.
func (orm *ORM) SaveBridgeType(bt *models.BridgeType) error {
	return orm.DB.Save(bt).Error
}

// SaveTx saves the transaction.
func (orm *ORM) SaveTx(tx *models.Tx) error {
	return orm.DB.Save(tx).Error
}

// SaveInitiator saves the initiator.
func (orm *ORM) SaveInitiator(initr *models.Initiator) error {
	return orm.DB.Save(initr).Error
}

// SaveHead saves the indexable block number related to head tracker.
func (orm *ORM) SaveHead(n *models.IndexableBlockNumber) error {
	return orm.DB.Save(n).Error
}

// LastHead returns the last ordered IndexableBlockNumber.
func (orm *ORM) LastHead() (*models.IndexableBlockNumber, error) {
	number := &models.IndexableBlockNumber{}
	rval := orm.DB.Order("digits desc, number desc").First(number)
	return number, multifyWithoutRecordNotFound(rval)
}

// DeleteStaleSessions deletes all sessions before the passed time.
func (orm *ORM) DeleteStaleSessions(before time.Time) error {
	return orm.DB.Where("last_used > ?", before).Delete(models.Session{}).Error
}

func (orm *ORM) SaveBulkDeleteRunTask(task *models.BulkDeleteRunTask) error {
	return orm.DB.Save(task).Error
}

func (orm *ORM) FindBulkDeleteRunTask(id string) (*models.BulkDeleteRunTask, error) {
	task := &models.BulkDeleteRunTask{}
	return task, orm.DB.Set("gorm:auto_preload", true).First(task, "ID = ?", id).Error
}
