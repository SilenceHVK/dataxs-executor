package executor

import (
	"context"
	"os"
	"path/filepath"

	"github.com/go-redis/redis/v8"
	"github.com/lexkong/log"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
)

type Job struct {
	JobName  string
	DataXEnv string
	Args     []string
	JobPath  string
	rdb      *redis.Client
}

func (e Job) Run() {
	log.Infof("开始执行同步任务 => 【%s】", e.JobName)
	context := context.Background()
	lastJobTime, errR := e.rdb.Get(context, e.JobName).Result()
	if errR != nil {
		lastJobTime = ""
	}

	var datax, err = Exec(context, e, append(e.Args, "-job", e.JobPath, "-lastJobTime", lastJobTime))
	if err != nil {
		log.Errorf(err, "")
		return
	}
	_ = datax.Wait()
}

func InitCronJob(dataxKey, jobKey, dataxEnv string) (*cron.Cron, error) {
	c := cron.New()

	// 设置 DataX 执行器参数
	var config Config
	if err := viper.UnmarshalKey(dataxKey, &config); err != nil {
		log.Error("解析 Config", err)
		return c, err
	}

	// 转换参数
	args, err := parseArgs(config)
	if err != nil {
		log.Error("转换参数 ", err)
		return c, err
	}

	// 连接 Redis
	var redisOption redis.Options
	if err := viper.UnmarshalKey("redis", &redisOption); err != nil {
		log.Error("读取 redis 配置错误", err)
		return c, err
	}
	rdb := redis.NewClient(&redisOption)

	// 获取程序执行目录
	rootPath, _ := os.Getwd()
	jobs := viper.GetStringMapStringSlice(jobKey)
	for spec, value := range jobs {
		for _, job := range value {
			log.Infof("添加同步任务 【%s】【%s】", spec, job)
			_ = c.AddJob(spec, &Job{
				JobName:  job,
				DataXEnv: dataxEnv,
				Args:     args,
				JobPath:  filepath.Join(rootPath, "job", job+config.JobExt),
				rdb:      rdb,
			})
		}
	}
	return c, nil
}
