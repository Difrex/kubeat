package beater

import (
	"encoding/json"
	"flag"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"errors"

	"io/ioutil"

	"strconv"
)

const (
	ELASTIC_ENV_USERNAME = "KUBEAT_ELASTIC_USERNAME"
	ELASTIC_ENV_PASSWORD = "KUBEAT_ELASTIC_PASSWORD"
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
	mux    sync.Mutex
}

type SenderConfig struct {
	Type     string   `json:"type"`
	Hosts    []string `json:"hosts"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Index    string   `json:"index"`
	DocType  string   `json:"doc_type"`
	Limit    int      `json:"limit"`
}

func GetSenderConfigFromFlags() *SenderConfig {
	sc := &SenderConfig{}
	path := flag.Lookup("sender-config").Value.String()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &sc)
	if err != nil {
		panic(err)
	}

	return sc
}

func GetTickFromFlags() int {
	tick := flag.Lookup("tick-time").Value.String()
	if tick != "" {
		i, err := strconv.Atoi(tick)
		if err != nil {
			panic(err)
		}
		return i
	}
	return 0
}

func isWatcherEnabled() bool {
	// watcher := flag.Lookup("enable-watcher").Value.String()
	// if watcher == "true" {
	// 	return true
	// }
	return false
}

type SenderClient interface {
	Connect(*SenderConfig) error
	Push(map[int64]LogMessage) error
}

func (p *PodLogs) NewSender() (err error) {
	var sender Sender
	var client SenderClient
	switch p.sc.Type {
	case "elasticsearch":
		e := &ElasticClient{}
		e.prefix = p.sc.Index
		e.docType = p.sc.DocType

		if p.sc.Username == "" || p.sc.Password == "" {
			p.sc.Username, p.sc.Password = getESCredsFromEnv()
		}

		client = SenderClient(e)
	case "tcp":
		c := &TCPClient{}
		c.conf = p.sc
		client = SenderClient(c)
	default:
		return errors.New("Wrong sender type")
	}

	if err := client.Connect(p.sc); err != nil {
		return err
	}

	sender.Client = client
	sender.box = newBox(p.sc)
	p.sender = &sender
	return
}

func (s *Sender) Send(ns, pod, message, con string) {
	l := LogMessage{
		Namespace:  ns,
		PodName:    pod,
		Message:    message,
		Container:  con,
		SenderTime: time.Now(),
	}

	s.add(l)

	if s.box.len >= s.box.limit {
		err := s.Client.Push(s.copyCon())
		if err != nil {
			log.Error(err)
			return
		}
		s.clean()
	}
}

func (s *Sender) copyCon() map[int64]LogMessage {
	s.mux.Lock()
	defer s.mux.Unlock()
	newMap := make(map[int64]LogMessage)
	for k, v := range s.box.con {
		newMap[k] = v
	}

	return newMap
}

func (s *Sender) Ticker() {
	ticker := time.NewTicker(time.Second * 60)
	for tick := range ticker.C {
		if s.len() > 0 {
			err := s.Client.Push(s.box.con)
			if err != nil {
				log.Error(err, " On tick ", tick.Unix())
			}
			s.clean()
		}
	}
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
	con   map[int64]LogMessage
	len   int
	limit int
	mux   sync.Mutex
}

func newBox(conf *SenderConfig) *box {
	return &box{
		con:   make(map[int64]LogMessage),
		limit: conf.Limit,
	}
}

func (s *Sender) len() int {
	s.box.len = len(s.box.con)
	return s.box.len
}

func (s *Sender) add(l LogMessage) {
	s.box.mux.Lock()
	defer s.box.mux.Unlock()
	s.box.con[time.Now().UnixNano()] = l
	s.box.len = len(s.box.con)
}

func (s *Sender) clean() {
	s.box.mux.Lock()
	defer s.box.mux.Unlock()
	s.box.con = make(map[int64]LogMessage)
	s.box.len = 0
}
