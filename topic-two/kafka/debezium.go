package kafka

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func (kc *KafkaCluster) CheckDebeziumConnector(url string, args ...string) (bool, error) {
	if len(args) != 0 {
		kc.DebeziumConnector = args[0]
	} else {
		kc.DebeziumConnector = "default_connector"
	}
	resp, err := http.Get(fmt.Sprintf("%s/%s", url, kc.DebeziumConnector))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return (resp.StatusCode == 200), nil
}

func (kc *KafkaCluster) CreateDebeziumConnector(url string) error {
	plan, err := os.ReadFile(fmt.Sprintf("./kafka/connector/%s.json", kc.DebeziumConnector))
	if err != nil {
		return err
	}
	io.Copy(log.Writer(), bytes.NewBuffer(plan))

	res, err := http.Post(fmt.Sprintf("%s/", url), "application/json", bytes.NewBuffer(plan))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	io.Copy(log.Writer(), res.Body)
	log.Println()
	if res.StatusCode != 200 {
		return fmt.Errorf("connector not added")
	}

	return nil
}
