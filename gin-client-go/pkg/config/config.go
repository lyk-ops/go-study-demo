package config

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
	"os"
	"strconv"
)

var keyMap map[KeyName]string

type Config struct {
	Server Server
}
type Server struct {
	Name string `yaml:"name"`
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

func init() {
	var config Config
	yamlFile, err := os.ReadFile("./.gin-client-go.yaml")
	if err != nil {
		klog.Fatal(err)
		return
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		klog.Fatal(err)
		return
	}
	keyMap = make(map[KeyName]string)
	keyMap[ServerName] = config.Server.Name
	keyMap[ServerHost] = config.Server.Host
	keyMap[ServerPort] = config.Server.Port
}
func GetString(keyName KeyName) string {
	return keyMap[keyName]
}

func GetInt(keyName KeyName) int {
	intStr := keyMap[keyName]
	fmt.Printf("intStr==>" + intStr + "")
	if intStr == "" {
		klog.Fatal("GetInt not read config==>" + keyName)
		return -1
	}
	v, err := strconv.Atoi(intStr)
	if err != nil {
		klog.Fatal(err)
		return -1
	}
	return v
}
