package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"time"
)

var config *viper.Viper

func init() {
	// 加载 config
	config = viper.New()
	config.SetConfigName("config")
	config.SetConfigType("yml")
	config.AddConfigPath("./conf")
	if err := config.ReadInConfig(); err != nil {
		panic(err)
	}
	config.WatchConfig()
	config.OnConfigChange(func(in fsnotify.Event) {
		fmt.Print("文件被更改了")
	})
	generateJobJson()
}

func generateJobJson() {

}

func main() {
	timer := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-timer.C:
			timer.Reset(10 * time.Second)
		}
	}
}
