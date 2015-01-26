package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Response struct {
	Servers []string `json:"servers"`
	Port    int      `json:"port"`
}

var (
	defaultEnvironmentFilePath = "/etc/zookeeper-hosts"
	defaultExhibitorURL        = "http://localhost:8181/exhibitor/v1/cluster/list"
	environmentFilePath        string
	exhibitorUrl               string
)

func init() {
	log.SetFlags(0)
	flag.StringVar(&environmentFilePath, "o", defaultEnvironmentFilePath, "environment file")
	flag.StringVar(&exhibitorUrl, "e", defaultExhibitorURL, "URL to exhibitor")
}

func main() {
	flag.Parse()
	tempFilePath := environmentFilePath + ".tmp"
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer tempFile.Close()
	if err := writeEnvironmentFile(tempFile); err != nil {
		log.Fatal(err)
	}
	os.Rename(tempFilePath, environmentFilePath)
}

func writeEnvironmentFile(w io.Writer) error {
	var buffer bytes.Buffer

	resp, err := http.Get(exhibitorUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var d Response
	err = json.Unmarshal(body, &d)
	if err != nil {
		return err
	}

	var tuples []string
	for _, server := range d.Servers {
		tuples = append(tuples, fmt.Sprintf("%s:%d", server, d.Port))
	}

	buffer.WriteString(fmt.Sprintf("ZK_SERVERS=%s\n", strings.Join(tuples, ",")))

	if _, err := buffer.WriteTo(w); err != nil {
		return err
	}

	return nil
}
