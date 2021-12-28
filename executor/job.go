package executor

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/lexkong/log"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"sync"
)

type Task struct {
	Name      string
	Successor []string
}
type Job struct {
	JobName  string
	DataXEnv string
	Args     []string
	JobPath  string
	JobExt   string
	rdb      *redis.Client
	Tasks    []Task
}

func (e *Job) Run() {
	for _, task := range e.Tasks {
		log.Infof("开始执行同步任务 => 【%s】", task.Name)
		if err := execAction(e, task.Name, nil); err != nil {
			log.Errorf(err, "执行同步任务 【%s】失败\n", task.Name)
			continue
		}

		// 开启协程执行后续操作
		var wg sync.WaitGroup
		wg.Add(len(task.Successor))
		for _, taskName := range task.Successor {
			go func(taskName string) {
				err := execAction(e, taskName, &wg)
				if err != nil {
					log.Errorf(err, "执行同步任务 【%s】失败", taskName)
				}
			}(taskName)
		}
		wg.Wait()
		// 执行任务成功后的后续任务
		log.Infof("同步任务【%s】执行成功", task.Name)
	}
}

func execAction(job *Job, taskName string, wg *sync.WaitGroup) error {
	_context := context.Background()
	job.JobName = taskName
	if wg != nil {
		defer wg.Done()
	}
	lastJobTime, errR := job.rdb.Get(_context, taskName).Result()
	if errR != nil {
		lastJobTime = ""
	}
	var datax, err = Exec(_context, job, append(job.Args, "-job", filepath.Join(job.JobPath, taskName+job.JobExt), "-lastJobTime", lastJobTime))
	if err != nil {
		log.Errorf(err, "执行同步任务 【%s】失败", taskName)
		return err
	}
	return datax.Wait()
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
	jobs := viper.GetStringMapString(jobKey)
	for spec := range jobs {
		log.Infof("添加同步任务 【%s】", spec)
		var tasks []Task
		if err := viper.UnmarshalKey(jobKey+"."+spec, &tasks); err != nil {
			log.Errorf(err, "读取 %s.%s 配置错误\n", jobKey, spec)
			continue
		}
		_ = c.AddJob(spec, &Job{
			DataXEnv: dataxEnv,
			Args:     args,
			JobPath:  filepath.Join(rootPath, "job"),
			JobExt:   config.JobExt,
			Tasks:    tasks,
			rdb:      rdb,
		})
	}
	return c, nil
}
