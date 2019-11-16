package store

import (
	"time"
	"utask/app"
	"utask/log"
	"utask/task"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

const (
	// 最大重试次数
	MaxRetryTimes = 32
	// 最大超时时间
	MaxLockTime = 300
)

// NewMysqlStore 返回一个新MysqlStore对象
var (
	// DefaultMysqlStore redis store
	DefaultMysqlStore = NewMysqlStore()
)

// NewMysqlStore 返回一个新MysqlStore对象
func NewMysqlStore() *MysqlStore {
	db, err := xorm.NewEngine(app.Config.Db.Driver, app.Config.Db.Source)
	if err != nil {
		log.Error("database err: ", err)
		panic(err)
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(30)
	db.SetConnMaxLifetime(20 * time.Minute) //默认30分钟的连接有效期
	db.ShowSQL(false)

	_ = db.Sync2(&TaskItem{}, &TaskLog{})
	return &MysqlStore{db}
}

// MysqlStore 是一个使用mysql实现的logStore，processStore
type MysqlStore struct {
	db *xorm.Engine
}

func (s *MysqlStore) Get(cid string, size int) (data []task.Tasker, err error) {
	log.Info("task process get: ", cid, size)
	m := make([]TaskItem, 0, size)

	lockTime := time.Now().Unix()
	nextLockTime := lockTime + app.Config.Db.MaxLockTime

	// 悲观获取
	rst, err := s.db.Exec(`UPDATE task_item
SET lock_status = 1, lock_time = ?, times = times + 1, cid = ?
WHERE cid = ? AND times < ? AND lock_time < ?
ORDER BY create_time ASC
LIMIT ?`, nextLockTime, cid, cid, app.Config.Db.MaxRetryTimes, lockTime, size)
	if err != nil {
		return nil, err
	}
	rows, err := rst.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, err
	}
	err = s.db.SQL(`SELECT * FROM task_item 
WHERE lock_status = 1 AND lock_time = ? AND cid = ?
ORDER BY create_time ASC
LIMIT ?`, nextLockTime, cid, size).Find(&m)
	if err != nil {
		return nil, err
	}
	for _, k := range m {
		item, err := Decode(k.Task)
		if err != nil {
			return nil, err
		}
		data = append(data, item)
	}
	return data, nil
}

func (s *MysqlStore) Insert(cid string, task task.Tasker) error {
	log.Info("task process insert: ", task)
	data, err := Encode(task)
	if err != nil {
		return err
	}
	lock := 0
	lockTime := task.GetNextTime()
	if task.IsProcessing() {
		lock = 1
		lockTime = time.Now().Unix() + task.Timeout()*2
	}
	_, err = s.db.Insert(&TaskItem{
		TID:        task.GetID(),
		AppID:      task.GetAppID(),
		Task:       data,
		Content:    task.GetContent(),
		Result:     "",
		Times:      0,
		LockTime:   lockTime,
		LockStatus: lock,
		SID:        task.GetSID(),
		CID:        cid,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	})
	return err
}

func (s *MysqlStore) Update(cid string, task task.Tasker) (bool, error) {
	log.Info("task process update: ", task)
	data, err := Encode(task)
	if err != nil {
		return false, err
	}
	errMsg := ""
	if err := task.GetLastError(); err != nil {
		errMsg = err.Error()
	}
	count, err := s.db.Update(&TaskItem{
		Task:       data,
		Result:     task.GetLastResult(),
		Error:      errMsg,
		ExecTime:   task.GetLastExecTime(),
		LockTime:   task.GetNextTime(),
		SID:        task.GetSID(),
		CID:        cid,
		UpdateTime: time.Now().Unix(),
	}, &TaskItem{
		TID: task.GetID(),
		CID: cid,
	})
	return count == 1, err
}

func (s *MysqlStore) Delete(cid string, id string) (bool, error) {
	log.Info("task process delete: ", id)
	count, err := s.db.Delete(&TaskItem{TID: id, CID: cid})
	return count == 1, err
}

func (s *MysqlStore) Log(cid string, task task.Tasker) error {
	log.Info("task log log: ", task)
	data, err := Encode(task)
	if err != nil {
		return err
	}
	errMsg := ""
	status := 1
	if err := task.GetLastError(); err != nil {
		errMsg = err.Error()
		status = 0
	}
	_, err = s.db.Insert(&TaskLog{TaskItem: TaskItem{
		TID:        task.GetID(),
		AppID:      task.GetAppID(),
		Task:       data,
		Content:    task.GetContent(),
		Result:     task.GetLastResult(),
		Error:      errMsg,
		ExecTime:   task.GetLastExecTime(),
		Times:      task.GetTimes(),
		LockTime:   task.GetNextTime(),
		SID:        task.GetSID(),
		CID:        cid,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}, Status: status})

	return err
}

type TaskItem struct {
	ID         int    `xorm:"'id' not null pk autoincr comment('自增ID') INT(11)"`
	TID        string `xorm:"'tid' not null comment('任务编号') index VARCHAR(36)"`
	AppID      string `xorm:"'app_id' not null comment('业务方ID') index VARCHAR(50)"`
	Task       []byte `xorm:"not null comment('任务') BLOB"`
	Content    string `xorm:"comment('任务内容') TEXT"`
	Result     string `xorm:"comment('任务结果') TEXT"`
	Error      string `xorm:"comment('错误信息') TEXT"`
	ExecTime   int64  `xorm:"not null comment('执行花费时间(毫秒)') INT(11)"`
	Times      int64  `xorm:"not null comment('执行次数') INT(11)"`
	LockTime   int64  `xorm:"comment('锁定时间戳') INT(11)"`
	LockStatus int    `xorm:"comment('锁定状态; 0:新建; 1:处理; 2: 完成;') TINYINT(4)"`
	SID        string `xorm:"'sid' not null comment('生产者ID') VARCHAR(36)"`
	CID        string `xorm:"'cid' not null comment('消费者ID') VARCHAR(36)"`
	CreateTime int64  `xorm:"not null comment('创建时间戳') INT(11)"`
	UpdateTime int64  `xorm:"not null comment('更新时间戳') INT(11)"`
}

type TaskLog struct {
	TaskItem `xorm:"extends"`
	Status   int `xorm:"comment('任务结果状态(1:成功, 0:失败)') TINYINT(4)"`
}
