package utils

import (
	"github.com/lexkong/log"
	"github.com/spf13/viper"
)

func Init(cfg string) error {
	c := &Config{Name: cfg}

	// 初始化配置文件
	if err := c.initConfig(); err != nil {
		return err
	}

	// 初始化日志
	if err := c.initLog(); err != nil {
		return err
	}
	return nil
}

type Config struct {
	Name string
}

// 初始化配置读取
func (c *Config) initConfig() error {
	if c.Name != "" {
		viper.SetConfigFile(c.Name)
	} else {
		viper.AddConfigPath("conf")
		viper.SetConfigName("config")
	}
	viper.SetConfigType("yaml")
	return viper.ReadInConfig()
}

// 初始化日志配置
func (c *Config) initLog() error {
	passLagerCfg := log.PassLagerCfg{
		Writers:        viper.GetString("log.writers"),
		LoggerLevel:    viper.GetString("log.logger_level"),
		LoggerFile:     viper.GetString("log.logger_file"),
		LogFormatText:  viper.GetBool("log.log_format_text"),
		RollingPolicy:  viper.GetString("log.rollingPolicy"),
		LogRotateDate:  viper.GetInt("log.log_rotate_date"),
		LogRotateSize:  viper.GetInt("log.log_rotate_size"),
		LogBackupCount: viper.GetInt("log.log_backup_count"),
	}
	return log.InitWithConfig(&passLagerCfg)
}
