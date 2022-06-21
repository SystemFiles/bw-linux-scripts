package secrets

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
)

const template = `{
  "api_key": "",
  "volume_uuid": ""
}
`

type Secrets struct {
	ApiKey     string `json:"api_key"`
	VolumeUUID string `json:"volume_uuid"`
}

func NewSecrets() Secrets {
	home, _ := os.UserHomeDir()
	configPath := path.Join(home, ".config/bw-snapshot")
	secretsFile := path.Join(configPath, "secrets.json")

	if fi, err := os.Stat(configPath); err != nil || !fi.IsDir() {
		err := os.MkdirAll(configPath, 0774)
		if err != nil {
			log.Fatalf("failed to create path to secrets file. %v", err)
		}
	}

	secretsContent, err := ioutil.ReadFile(secretsFile)
	if err != nil {
		if f, err := os.Create(secretsFile); err == nil {
			log.Println("creating secrets file since one did not exist ... add your secrets before running again!")
			f.Write([]byte(template))
			f.Close()
		}
		os.Exit(1)
	}

	var sj Secrets
	err = json.Unmarshal(secretsContent, &sj)
	if err != nil {
		log.Fatalf("failed to parse secrets file. %v", err)
	}

	if sj.ApiKey == "" || sj.VolumeUUID == "" {
		log.Fatalf("'api_key' or 'volume_uuid' was not provided in secrets file. edit %s and try again", secretsFile)
	}

	return sj
}