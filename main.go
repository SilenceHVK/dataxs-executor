package main

import (
	"os/exec"

	"github.com/SilenceHVK/dataxs-executor/executor"

	"github.com/gin-gonic/gin"

	"github.com/SilenceHVK/dataxs-executor/utils"
	"github.com/lexkong/log"
	"github.com/spf13/viper"
)

func main() {
	// 加载配置文件
	if err := utils.Init(""); err != nil {
		panic(err)
	}

	// 获取 Datax 执行器环境
	dataxEnv := viper.GetString("datax.env")
	if _, err := exec.LookPath(dataxEnv); err != nil {
		log.Warnf("Datax 依赖 %s 环境未找到，请安装....", dataxEnv)
		return
	}

	// 初始化 CronJob
	c, err := executor.InitCronJob("datax", "jobs", dataxEnv)
	if err != nil {
		log.Errorf(err, "")
		return
	}

	c.Start()
	defer c.Stop()
	log.Info("==================== 同步应用启动 ====================")

	gin.SetMode(viper.GetString("server.mode"))
	r := gin.Default()
	_ = r.Run(":" + viper.GetString("server.port"))
}
