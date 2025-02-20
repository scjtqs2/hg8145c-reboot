package main

import (
	"context"
	"github.com/robfig/cron/v3"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"google.golang.org/appengine/log"
	"net/http"
	"os"
	"time"
)

type Job struct {
	wd       selenium.WebDriver
	seAddr   string // selenium 服务的地址
	cron     *cron.Cron
	loginUrl string // 光猫的web登录页地址
	username string // 光猫的web用户名，非管理员的，在光猫的底部贴纸上
	password string // 光猫的web密码，非管理员的，在光猫的底部贴纸上
	crontab  string // 定时任务的计时规则。和linux的crontab的规则一致
}

// NewJob 初始化Job类
func NewJob() *Job {
	se := &Job{
		seAddr:   os.Getenv("SELENIUM_ADDR"),
		loginUrl: os.Getenv("LOGIN_URL"),
		username: os.Getenv("USERNAME"),
		password: os.Getenv("PASSWORD"),
		crontab:  os.Getenv("CRONTAB"),
	}
	// 定时任务开启
	se.cron = cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))
	return se
}

// newChrome 初始化chrome的webdriver
func (g *Job) newChrome() (selenium.WebDriver, error) {
	selenium.HTTPClient = &http.Client{
		Timeout: time.Second * 30,
	}
	caps := selenium.Capabilities{"browserName": "chrome"}
	// chrome参数
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--headless", // 设置Chrome无头模式，在linux下运行，需要设置这个参数，否则会报错
			"--disable-gpu",
			// "--no-sandbox",
			"--window-size=1366,768",
			// fmt.Sprintf("--proxy-server=%s", "http://192.168.28.101:7890"), // --proxy-server=http://127.0.0.1:1234
		},
		W3C: true,
	}
	caps.AddChrome(chromeCaps)
	wd, err := selenium.NewRemote(caps, g.seAddr)
	return wd, err
}

// login 登录光猫管理界面
func (g *Job) login(username string, password string) error {
	if err := g.wd.Get(g.seAddr); err != nil {
		return err
	}
	_ = g.wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		_, err := wd.FindElement(selenium.ByCSSSelector, "#user_name")
		return err == nil, nil
	})
	usenameRequest, err := g.wd.FindElement(selenium.ByCSSSelector, "#user_name")
	if err != nil {
		return err
	}
	usenameRequest.Clear()
	usenameRequest.SendKeys(username)
	passwordRequest, err := g.wd.FindElement(selenium.ByCSSSelector, "#password")
	if err != nil {
		return err
	}
	passwordRequest.SendKeys(password)
	submitBtn, err := g.wd.FindElement(selenium.ByCSSSelector, "#save")
	if err != nil {
		return err
	}
	err = submitBtn.Click()
	if err != nil {
		return err
	}
	return nil
}

// switchToReboot 切换到重启界面并点击重启按钮
func (g *Job) switchToReboot() error {
	// 进入 “管理”界面
	err := g.wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		_, err := g.wd.FindElement(selenium.ByCSSSelector, "#Menu1_Managemen > div.item_link > a")
		return err == nil, nil
	})
	if err != nil {
		return err
	}
	guanli, err := g.wd.FindElement(selenium.ByCSSSelector, "#Menu1_Managemen > div.item_link > a")
	if err != nil {
		return err
	}
	err = guanli.Click()
	if err != nil {
		return err
	}
	// 进入“设备”界面
	err = g.wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		_, err = g.wd.FindElement(selenium.ByCSSSelector, "#Menu2_Mng_Device > a")
		return err == nil, nil
	})
	if err != nil {
		return err
	}
	shebei, err := g.wd.FindElement(selenium.ByCSSSelector, "#Menu2_Mng_Device > a")
	if err != nil {
		return err
	}
	err = shebei.Click()
	if err != nil {
		return err
	}
	// 需要切换iframe
	err = g.wd.SwitchFrame("frameContent")
	if err != nil {
		return err
	}
	// 找到“重启”按钮
	err = g.wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		_, err = g.wd.FindElement(selenium.ByCSSSelector, "#Restart_button")
		return err == nil, nil
	})
	if err != nil {
		return err
	}
	chongqi, err := g.wd.FindElement(selenium.ByCSSSelector, "#Restart_button")
	if err != nil {
		return err
	}
	// 点击“重启”按钮
	err = chongqi.Click()
	if err != nil {
		return err
	}
	// “确认”alert弹窗
	return g.wd.AcceptAlert()
}

// screenShort 将当前画面记录到本地
func (g *Job) screenShort() error {
	b, err := g.wd.Screenshot()
	if err != nil {
		return err
	}
	return os.WriteFile("tmp.png", b, 0777)
}

// pageSource 将当前页面的源码保存到本地
func (g *Job) pageSource() error {
	b, err := g.wd.PageSource()
	if err != nil {
		return err
	}
	return os.WriteFile("index.html", []byte(b), 0777)
}

// reboot 重启流程
func (g *Job) reboot() error {
	var err error
	g.wd, err = g.newChrome()
	if err != nil {
		return err
	}
	defer g.wd.Quit()
	defer g.pageSource()
	defer g.screenShort()
	err = g.login(g.username, g.password)
	if err != nil {
		return err
	}
	return g.switchToReboot()
}

func (g *Job) execJob() {
	ctx := context.TODO()
	err := g.reboot()
	if err != nil {
		log.Errorf(ctx, "reboot faild %v", err)
	}
}

// main 主入口
func main() {
	var err error
	se := NewJob()
	_, err = se.cron.AddFunc(se.crontab, se.execJob)
	if err != nil {
		panic(err)
	}
	se.cron.Run()
}
