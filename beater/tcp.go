package beater

import (
	"encoding/json"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

type TCPClient struct {
	Client net.Conn
	conf   *SenderConfig
}

func (t *TCPClient) Connect(conf *SenderConfig) (err error) {
	conn, err := net.Dial("tcp", strings.Join(conf.Hosts, ""))
	t.Client = conn
	return
}

func (t *TCPClient) Push(l map[int64]LogMessage) error {
	log.Infof("Tying send %d logs", len(l))
	for _, v := range l {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}

		d := string(data) + "\n"
		if _, err := t.Client.Write([]byte(d)); err != nil {
			panic(err)
		}
	}
	return nil
}
