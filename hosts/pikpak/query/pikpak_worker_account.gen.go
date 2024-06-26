// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
)

func newWorkerAccount(db *gorm.DB, opts ...gen.DOOption) workerAccount {
	_workerAccount := workerAccount{}

	_workerAccount.workerAccountDo.UseDB(db, opts...)
	_workerAccount.workerAccountDo.UseModel(&model.WorkerAccount{})

	tableName := _workerAccount.workerAccountDo.TableName()
	_workerAccount.ALL = field.NewAsterisk(tableName)
	_workerAccount.UserID = field.NewString(tableName, "user_id")
	_workerAccount.MasterUserID = field.NewString(tableName, "master_user_id")
	_workerAccount.Email = field.NewString(tableName, "email")
	_workerAccount.Password = field.NewString(tableName, "password")
	_workerAccount.UsedSize = field.NewInt64(tableName, "used_size")
	_workerAccount.LimitSize = field.NewInt64(tableName, "limit_size")
	_workerAccount.PremiumExpiration = field.NewTime(tableName, "premium_expiration")
	_workerAccount.CreatedAt = field.NewTime(tableName, "created_at")
	_workerAccount.UpdatedAt = field.NewTime(tableName, "updated_at")
	_workerAccount.InvalidUntil = field.NewTime(tableName, "invalid_until")
	_workerAccount.UpdatedUUID = field.NewString(tableName, "updated_uuid")

	_workerAccount.fillFieldMap()

	return _workerAccount
}

type workerAccount struct {
	workerAccountDo

	ALL               field.Asterisk
	UserID            field.String
	MasterUserID      field.String
	Email             field.String
	Password          field.String
	UsedSize          field.Int64
	LimitSize         field.Int64
	PremiumExpiration field.Time
	CreatedAt         field.Time
	UpdatedAt         field.Time
	InvalidUntil      field.Time
	UpdatedUUID       field.String

	fieldMap map[string]field.Expr
}

func (w workerAccount) Table(newTableName string) *workerAccount {
	w.workerAccountDo.UseTable(newTableName)
	return w.updateTableName(newTableName)
}

func (w workerAccount) As(alias string) *workerAccount {
	w.workerAccountDo.DO = *(w.workerAccountDo.As(alias).(*gen.DO))
	return w.updateTableName(alias)
}

func (w *workerAccount) updateTableName(table string) *workerAccount {
	w.ALL = field.NewAsterisk(table)
	w.UserID = field.NewString(table, "user_id")
	w.MasterUserID = field.NewString(table, "master_user_id")
	w.Email = field.NewString(table, "email")
	w.Password = field.NewString(table, "password")
	w.UsedSize = field.NewInt64(table, "used_size")
	w.LimitSize = field.NewInt64(table, "limit_size")
	w.PremiumExpiration = field.NewTime(table, "premium_expiration")
	w.CreatedAt = field.NewTime(table, "created_at")
	w.UpdatedAt = field.NewTime(table, "updated_at")
	w.InvalidUntil = field.NewTime(table, "invalid_until")
	w.UpdatedUUID = field.NewString(table, "updated_uuid")

	w.fillFieldMap()

	return w
}

func (w *workerAccount) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := w.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (w *workerAccount) fillFieldMap() {
	w.fieldMap = make(map[string]field.Expr, 11)
	w.fieldMap["user_id"] = w.UserID
	w.fieldMap["master_user_id"] = w.MasterUserID
	w.fieldMap["email"] = w.Email
	w.fieldMap["password"] = w.Password
	w.fieldMap["used_size"] = w.UsedSize
	w.fieldMap["limit_size"] = w.LimitSize
	w.fieldMap["premium_expiration"] = w.PremiumExpiration
	w.fieldMap["created_at"] = w.CreatedAt
	w.fieldMap["updated_at"] = w.UpdatedAt
	w.fieldMap["invalid_until"] = w.InvalidUntil
	w.fieldMap["updated_uuid"] = w.UpdatedUUID
}

func (w workerAccount) clone(db *gorm.DB) workerAccount {
	w.workerAccountDo.ReplaceConnPool(db.Statement.ConnPool)
	return w
}

func (w workerAccount) replaceDB(db *gorm.DB) workerAccount {
	w.workerAccountDo.ReplaceDB(db)
	return w
}

type workerAccountDo struct{ gen.DO }

type IWorkerAccountDo interface {
	gen.SubQuery
	Debug() IWorkerAccountDo
	WithContext(ctx context.Context) IWorkerAccountDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IWorkerAccountDo
	WriteDB() IWorkerAccountDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IWorkerAccountDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IWorkerAccountDo
	Not(conds ...gen.Condition) IWorkerAccountDo
	Or(conds ...gen.Condition) IWorkerAccountDo
	Select(conds ...field.Expr) IWorkerAccountDo
	Where(conds ...gen.Condition) IWorkerAccountDo
	Order(conds ...field.Expr) IWorkerAccountDo
	Distinct(cols ...field.Expr) IWorkerAccountDo
	Omit(cols ...field.Expr) IWorkerAccountDo
	Join(table schema.Tabler, on ...field.Expr) IWorkerAccountDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IWorkerAccountDo
	RightJoin(table schema.Tabler, on ...field.Expr) IWorkerAccountDo
	Group(cols ...field.Expr) IWorkerAccountDo
	Having(conds ...gen.Condition) IWorkerAccountDo
	Limit(limit int) IWorkerAccountDo
	Offset(offset int) IWorkerAccountDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IWorkerAccountDo
	Unscoped() IWorkerAccountDo
	Create(values ...*model.WorkerAccount) error
	CreateInBatches(values []*model.WorkerAccount, batchSize int) error
	Save(values ...*model.WorkerAccount) error
	First() (*model.WorkerAccount, error)
	Take() (*model.WorkerAccount, error)
	Last() (*model.WorkerAccount, error)
	Find() ([]*model.WorkerAccount, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.WorkerAccount, err error)
	FindInBatches(result *[]*model.WorkerAccount, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.WorkerAccount) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IWorkerAccountDo
	Assign(attrs ...field.AssignExpr) IWorkerAccountDo
	Joins(fields ...field.RelationField) IWorkerAccountDo
	Preload(fields ...field.RelationField) IWorkerAccountDo
	FirstOrInit() (*model.WorkerAccount, error)
	FirstOrCreate() (*model.WorkerAccount, error)
	FindByPage(offset int, limit int) (result []*model.WorkerAccount, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IWorkerAccountDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (w workerAccountDo) Debug() IWorkerAccountDo {
	return w.withDO(w.DO.Debug())
}

func (w workerAccountDo) WithContext(ctx context.Context) IWorkerAccountDo {
	return w.withDO(w.DO.WithContext(ctx))
}

func (w workerAccountDo) ReadDB() IWorkerAccountDo {
	return w.Clauses(dbresolver.Read)
}

func (w workerAccountDo) WriteDB() IWorkerAccountDo {
	return w.Clauses(dbresolver.Write)
}

func (w workerAccountDo) Session(config *gorm.Session) IWorkerAccountDo {
	return w.withDO(w.DO.Session(config))
}

func (w workerAccountDo) Clauses(conds ...clause.Expression) IWorkerAccountDo {
	return w.withDO(w.DO.Clauses(conds...))
}

func (w workerAccountDo) Returning(value interface{}, columns ...string) IWorkerAccountDo {
	return w.withDO(w.DO.Returning(value, columns...))
}

func (w workerAccountDo) Not(conds ...gen.Condition) IWorkerAccountDo {
	return w.withDO(w.DO.Not(conds...))
}

func (w workerAccountDo) Or(conds ...gen.Condition) IWorkerAccountDo {
	return w.withDO(w.DO.Or(conds...))
}

func (w workerAccountDo) Select(conds ...field.Expr) IWorkerAccountDo {
	return w.withDO(w.DO.Select(conds...))
}

func (w workerAccountDo) Where(conds ...gen.Condition) IWorkerAccountDo {
	return w.withDO(w.DO.Where(conds...))
}

func (w workerAccountDo) Exists(subquery interface{ UnderlyingDB() *gorm.DB }) IWorkerAccountDo {
	return w.Where(field.CompareSubQuery(field.ExistsOp, nil, subquery.UnderlyingDB()))
}

func (w workerAccountDo) Order(conds ...field.Expr) IWorkerAccountDo {
	return w.withDO(w.DO.Order(conds...))
}

func (w workerAccountDo) Distinct(cols ...field.Expr) IWorkerAccountDo {
	return w.withDO(w.DO.Distinct(cols...))
}

func (w workerAccountDo) Omit(cols ...field.Expr) IWorkerAccountDo {
	return w.withDO(w.DO.Omit(cols...))
}

func (w workerAccountDo) Join(table schema.Tabler, on ...field.Expr) IWorkerAccountDo {
	return w.withDO(w.DO.Join(table, on...))
}

func (w workerAccountDo) LeftJoin(table schema.Tabler, on ...field.Expr) IWorkerAccountDo {
	return w.withDO(w.DO.LeftJoin(table, on...))
}

func (w workerAccountDo) RightJoin(table schema.Tabler, on ...field.Expr) IWorkerAccountDo {
	return w.withDO(w.DO.RightJoin(table, on...))
}

func (w workerAccountDo) Group(cols ...field.Expr) IWorkerAccountDo {
	return w.withDO(w.DO.Group(cols...))
}

func (w workerAccountDo) Having(conds ...gen.Condition) IWorkerAccountDo {
	return w.withDO(w.DO.Having(conds...))
}

func (w workerAccountDo) Limit(limit int) IWorkerAccountDo {
	return w.withDO(w.DO.Limit(limit))
}

func (w workerAccountDo) Offset(offset int) IWorkerAccountDo {
	return w.withDO(w.DO.Offset(offset))
}

func (w workerAccountDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IWorkerAccountDo {
	return w.withDO(w.DO.Scopes(funcs...))
}

func (w workerAccountDo) Unscoped() IWorkerAccountDo {
	return w.withDO(w.DO.Unscoped())
}

func (w workerAccountDo) Create(values ...*model.WorkerAccount) error {
	if len(values) == 0 {
		return nil
	}
	return w.DO.Create(values)
}

func (w workerAccountDo) CreateInBatches(values []*model.WorkerAccount, batchSize int) error {
	return w.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (w workerAccountDo) Save(values ...*model.WorkerAccount) error {
	if len(values) == 0 {
		return nil
	}
	return w.DO.Save(values)
}

func (w workerAccountDo) First() (*model.WorkerAccount, error) {
	if result, err := w.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.WorkerAccount), nil
	}
}

func (w workerAccountDo) Take() (*model.WorkerAccount, error) {
	if result, err := w.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.WorkerAccount), nil
	}
}

func (w workerAccountDo) Last() (*model.WorkerAccount, error) {
	if result, err := w.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.WorkerAccount), nil
	}
}

func (w workerAccountDo) Find() ([]*model.WorkerAccount, error) {
	result, err := w.DO.Find()
	return result.([]*model.WorkerAccount), err
}

func (w workerAccountDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.WorkerAccount, err error) {
	buf := make([]*model.WorkerAccount, 0, batchSize)
	err = w.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (w workerAccountDo) FindInBatches(result *[]*model.WorkerAccount, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return w.DO.FindInBatches(result, batchSize, fc)
}

func (w workerAccountDo) Attrs(attrs ...field.AssignExpr) IWorkerAccountDo {
	return w.withDO(w.DO.Attrs(attrs...))
}

func (w workerAccountDo) Assign(attrs ...field.AssignExpr) IWorkerAccountDo {
	return w.withDO(w.DO.Assign(attrs...))
}

func (w workerAccountDo) Joins(fields ...field.RelationField) IWorkerAccountDo {
	for _, _f := range fields {
		w = *w.withDO(w.DO.Joins(_f))
	}
	return &w
}

func (w workerAccountDo) Preload(fields ...field.RelationField) IWorkerAccountDo {
	for _, _f := range fields {
		w = *w.withDO(w.DO.Preload(_f))
	}
	return &w
}

func (w workerAccountDo) FirstOrInit() (*model.WorkerAccount, error) {
	if result, err := w.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.WorkerAccount), nil
	}
}

func (w workerAccountDo) FirstOrCreate() (*model.WorkerAccount, error) {
	if result, err := w.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.WorkerAccount), nil
	}
}

func (w workerAccountDo) FindByPage(offset int, limit int) (result []*model.WorkerAccount, count int64, err error) {
	result, err = w.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = w.Offset(-1).Limit(-1).Count()
	return
}

func (w workerAccountDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = w.Count()
	if err != nil {
		return
	}

	err = w.Offset(offset).Limit(limit).Scan(result)
	return
}

func (w workerAccountDo) Scan(result interface{}) (err error) {
	return w.DO.Scan(result)
}

func (w workerAccountDo) Delete(models ...*model.WorkerAccount) (result gen.ResultInfo, err error) {
	return w.DO.Delete(models)
}

func (w *workerAccountDo) withDO(do gen.Dao) *workerAccountDo {
	w.DO = *do.(*gen.DO)
	return w
}
