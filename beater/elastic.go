package beater

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"github.com/vjeantet/jodaTime"
)

type ElasticClient struct {
	Client  *elastic.Client
	docType string
	prefix  string
}

func (e *ElasticClient) Connect(conf *SenderConfig) (err error) {
	var client *elastic.Client
	if conf.Username != "" && conf.Password != "" {
		client, err = elastic.NewClient(
			elastic.SetURL(conf.Hosts...),
			elastic.SetSniff(false),
			elastic.SetBasicAuth(conf.Username, conf.Password))
	} else {
		client, err = elastic.NewClient(
			elastic.SetSniff(false),
			elastic.SetURL(conf.Hosts...))
	}
	e.Client = client
	return
}

func (e *ElasticClient) checkIndex() (bool, error) {
	if e.Client == nil {
		return false, errors.New("ElasticSearch client not initialized")
	}

	if ok, err := e.Client.IndexExists(e.indexName()).Do(context.Background()); !ok {
		return ok, err
	}
	return true, nil
}

// TODO: support for different index patterns
func (e *ElasticClient) createIndex() error {
	if e.Client == nil {
		return errors.New("ElasticSearch client not initialized")
	}

	_, err := e.Client.CreateIndex(e.indexName()).Do(context.Background())
	return err
}

func (e *ElasticClient) indexName() string {
	return e.prefix + "-" + jodaTime.Format("YYYY.MM.dd", time.Now())
}

func (e *ElasticClient) Push(l map[int64]LogMessage) error {
	if ok, err := e.checkIndex(); !ok && err != nil {
		return err
	} else if !ok {
		if err := e.createIndex(); err != nil {
			return err
		}
	}

	bulk := e.Client.Bulk()
	for _, v := range l {
		r := elastic.NewBulkIndexRequest().
			Index(e.indexName()).
			Type(e.docType).
			Id(uuid.New().String()).
			Doc(v)
		bulk = bulk.Add(r)
	}
	log.Infof("Sending %d messages to the ElasticSearch", len(l))
	resp, err := bulk.Do(context.Background())
	log.Infof("Indexed. Took %d", resp.Took)

	// Refresh index
	// e.Client.Index().Refresh(e.indexName())

	return err
}
