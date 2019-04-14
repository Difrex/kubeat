package beater

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/vjeantet/jodaTime"

	"errors"

	"context"

	"github.com/google/uuid"
	"github.com/olivere/elastic"
)

type LogMessage struct {
	PodName    string                 `json:"pod_name"`
	Namespace  string                 `json:"namespace"`
	Container  string                 `json:"container"`
	Message    string                 `json:"message"`
	SenderTime time.Time              `json:"sender_time"`
	Meta       map[string]interface{} `json:"meta"`
}

type Sender struct {
	Client SenderClient
	Config *SenderConfig
	box    *box
}

type SenderConfig struct {
	Type     string   `json:"type"`
	Hosts    []string `json:"hosts"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Index    string   `json:"string"`
	DocType  string   `json:"doc_type"`
}

type SenderClient interface {
	Connect(*SenderConfig) error
	Push([]LogMessage) error
}

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
			elastic.SetBasicAuth(conf.Username, conf.Password))
	} else {
		client, err = elastic.NewClient(elastic.SetURL(conf.Hosts...))
	}
	e.Client = client
	return
}

func (e *ElasticClient) checkIndex(conf *SenderConfig) (bool, error) {
	if e.Client == nil {
		return false, errors.New("ElasticSearch client not initialized")
	}

	if ok, err := e.Client.IndexExists(conf.Index).Do(context.Background()); !ok {
		return false, err
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

func (e *ElasticClient) Push(l []LogMessage) error {
	e.Client.Index().
		Index(e.indexName()).
		Id(uuid.New().String()).
		Type(e.docType).
		BodyJson(l).
		Do(context.Background())
	return nil
}

func (p *PodLogs) NewSender() (err error) {
	var sender Sender
	var client SenderClient
	switch p.sc.Type {
	case "elasticsearch":
		client = SenderClient(&ElasticClient{})
		if err := client.Connect(p.sc); err != nil {
			return err
		}
	default:
		return errors.New("Wrong sender type")
	}

	sender.Client = client
	p.sender = &sender
	return
}

func logSender(ns, pod, message, con string) error {
	log := LogMessage{
		Namespace:  ns,
		PodName:    pod,
		Message:    message,
		Container:  con,
		SenderTime: time.Now(),
	}
	data, err := json.Marshal(log)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

type box struct {
	con []LogMessage
	mux sync.Mutex
	len int
}

func (c *box) Len() int {
	c.mux.Lock()
	c.len = len(c.con)
	c.mux.Unlock()
	return c.len
}

func (c *box) Add(l LogMessage) {
	c.mux.Lock()
	c.con = append(c.con, l)
	c.mux.Unlock()
}

func (c *box) Clean() {
	c.mux.Lock()
	c.con = []LogMessage{}
	c.mux.Unlock()
}
