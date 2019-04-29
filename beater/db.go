package beater

import (
	memdb "github.com/hashicorp/go-memdb"
	log "github.com/sirupsen/logrus"
)

type LogWatcher struct {
	Name string
	Chan chan bool
}

func NewDB() (*memdb.MemDB, error) {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"logwatchers": &memdb.TableSchema{
				Name: "logwatchers",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
				},
			},
			"logmessage": &memdb.TableSchema{
				Name: "logmessage",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Id"},
					},
				},
			},
		},
	}

	return memdb.NewMemDB(schema)
}

func (p *PodLogs) AddWatcherToDb(pod string) (chan bool, error) {
	txn := p.db.Txn(true)

	ch := make(chan bool, 1)
	watcher := &LogWatcher{
		Name: pod,
		Chan: ch,
	}

	err := txn.Insert("logwatchers", watcher)
	txn.Commit()

	return ch, err
}

func (p *PodLogs) GetWatcherFromDB(pod string) (*LogWatcher, error) {
	txn := p.db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("logwatchers", "id", pod)
	if err != nil {
		return nil, err
	}

	if raw == nil {
		return nil, nil
	}

	return raw.(*LogWatcher), err
}

func (p *PodLogs) DelWatcherFromDB(pod string) error {
	log.Debugf("Trying delete pod %s from DB", pod)
	watcher, err := p.GetWatcherFromDB(pod)
	if err != nil {
		return err
	}

	if watcher == nil {
		return nil
	}

	txn := p.db.Txn(true)
	if err := txn.Delete("logwatchers", watcher); err != nil {
		return err
	}
	txn.Commit()

	return nil
}

func (p *PodLogs) IsWatcherInTheDB(pod string) (bool, *LogWatcher, error) {
	watcher, err := p.GetWatcherFromDB(pod)
	if err != nil {
		return false, nil, err
	}

	if watcher != nil {
		return true, watcher, nil
	}

	return false, nil, nil
}

func (p *PodLogs) GetWatchersFromDBLen() int {
	txn := p.db.Txn(false)
	defer txn.Abort()

	i, err := txn.Get("logwatchers", "id")
	if err != nil {
		log.Error(err)
		return -1
	}

	var c int
	for item := i.Next(); item != nil; item = i.Next() {
		c++
	}

	return c
}
