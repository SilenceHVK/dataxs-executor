package executor

import (
	"bufio"
	"context"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lexkong/log"
)

type Config struct {
	Xms       string
	Xmx       string
	JobExt    string
	Mode      string
	LogLevel  string
	DataXHome string
	JobId     int
}

type DataX struct {
	Tag     string
	Command *exec.Cmd
	Pid     int
	parent  *context.Context
	rdb     *redis.Client
}

func (d DataX) Wait() (err error) {
	log.Infof("【%s】任务执行中 PID ====> 【%d】", d.Tag, d.Pid)
	err = d.Command.Wait()
	if d.Command.ProcessState != nil {
		log.Infof("【%s】任务执行完成 PID ====> 【%d】", d.Tag, d.Pid)
		// 只有正常结束的任务才记录时间
		if d.Command.ProcessState.ExitCode() == 0 {
			_ = d.rdb.Set(*d.parent, d.Tag, time.Now().Format("2006-01-02 15:04:05"), 0).Err()
		}
	}
	return err
}

func Exec(ctx context.Context, job *Job, args []string) (datax *DataX, err error) {
	command := exec.CommandContext(ctx, job.DataXEnv, args...)
	stderr, _ := command.StderrPipe()
	stdout, _ := command.StdoutPipe()

	datax = &DataX{
		Tag:     job.JobName,
		Command: command,
		parent:  &ctx,
		rdb:     job.rdb,
	}

	go func() {
		defer stderr.Close()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Infof("【%s】 ----> %s", job.JobName, scanner.Text())
		}
	}()

	go func() {
		defer stdout.Close()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.Infof("【%s】 ----> %s", job.JobName, scanner.Text())
		}
	}()

	err = command.Start()
	if err != nil {
		log.Errorf(err, "")
		return nil, err
	}
	datax.Pid = command.Process.Pid
	return
}

func parseArgs(config Config) ([]string, error) {
	dataxHome, err := filepath.Abs(config.DataXHome)
	if err != nil {
		log.Errorf(err, "获取 DataX 目录失败")
		return nil, err
	}

	args := []string{
		"-server",
		"-Xms" + config.Xms,
		"-Xmx" + config.Xmx,
		"-XX:+HeapDumpOnOutOfMemoryError",
		"-XX:HeapDumpPath=" + filepath.Join(dataxHome, "log"),
		"-Dloglevel=" + config.LogLevel,
		"-Dfile.encoding=UTF-8",
		"-Dlogback.statusListenerClass=ch.qos.logback.core.status.NopStatusListener",
		"-Djava.security.egd=file:///dev/urandom",
		"-Ddatax.home=" + dataxHome,
		"-Dlogback.configurationFile=" + filepath.Join(dataxHome, "conf/logback.xml"),
		"-classpath",
		filepath.Join(dataxHome, "lib/*:."),
		"-Dlog.file.name=dlog_" + strconv.Itoa(config.JobId),
		"com.alibaba.datax.core.Engine",
		"-mode",
		config.Mode,
		"-jobid",
		strconv.Itoa(config.JobId),
	}

	return args, nil
}
