package config

import (
	"hash/fnv"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type SingleURL struct {
	Name    string `yaml:"name"`
	Url     string `yaml:"url"`
	UserID  int64  `yaml:"user_id"`
	Enabled bool   `yaml:"enabled"`
	Type    string `yaml:"type"`
}

type ElementIds struct {
	CaptchaID           string `yaml:"captcha_id"`
	CaptchaInputID      string `yaml:"captcha_txt_input_id"`
	CaptchaErrID        string `yaml:"captcha_err_id"`
	CaptchaButtonID     string `yaml:"captcha_button_id"`
	SecondStageButtonID string `yaml:"second_stage_button_id"`

	SecondStageBisCheckID  string `yaml:"second_stage_bis_check_id"`
	SecondStageBisButtonID string `yaml:"second_stage_bis_button_id"`
}

type Config struct {
	ApiKey                 string       `yaml:"api_key"`
	Port                   int32        `yaml:"port"`
	IntervalMinute         int          `yaml:"minute_interval"`
	AllowedRequestsMinHour int          `yaml:"allowed_requests_min_hour"`
	AllowedRequestsMaxHour int          `yaml:"allowed_requests_max_hour"`
	HealthCheckMinHour     int          `yaml:"health_check_min_hour"`
	HealthCheckMaxHour     int          `yaml:"health_check_max_hour"`
	HtmlElems              ElementIds   `yaml:"html"`
	Urls                   []*SingleURL `yaml:"urls"`
}

func ReadConfig(cfg *Config) {
	f, err := os.Open("configs/config.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		panic(err)
	}

	if cfg.AllowedRequestsMinHour > cfg.HealthCheckMinHour ||
		cfg.AllowedRequestsMaxHour < cfg.HealthCheckMaxHour {
		panic("Health check interval must be inside allowed request hour interval")
	}
}

func (r SingleURL) GetHash() uint32 {
	h := fnv.New32a()
	h.Write([]byte(r.Url + r.Name + strconv.Itoa(int(r.UserID))))
	return h.Sum32()
}
