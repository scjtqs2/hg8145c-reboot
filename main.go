package main

import (
	"context"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
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
	ctx      context.Context
}

// NewJob 初始化Job类
func NewJob() *Job {
	se := &Job{
		seAddr:   os.Getenv("SELENIUM_ADDR"),
		loginUrl: os.Getenv("LOGIN_URL"),
		username: os.Getenv("LOGIN_USERNAME"),
		password: os.Getenv("LOGIN_PASSWORD"),
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
	if err := g.wd.Get(g.loginUrl); err != nil {
		return err
	}
	_ = g.wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		_, err := wd.FindElement(selenium.ByCSSSelector, "#user_name")
		return err == nil, nil
	})
	usenameRequest, err := g.wd.FindElement(selenium.ByCSSSelector, "#user_name")
	if err != nil {
		log.Errorf("查找用户名输入框失败 err=%v", err)
		return err
	}
	err = usenameRequest.Clear() // 清空预制的密码。
	if err != nil {
		log.Errorf("清空用户名输入框内容失败 err=%v", err)
		return err
	}
	err = usenameRequest.SendKeys(username)
	if err != nil {
		log.Errorf("输入用户名失败 err=%v", err)
		return err
	}
	passwordRequest, err := g.wd.FindElement(selenium.ByCSSSelector, "#password")
	if err != nil {
		log.Errorf("查找密码输入框失败 err=%v", err)
		return err
	}
	err = passwordRequest.SendKeys(password)
	if err != nil {
		log.Errorf("输入密码失败 err=%v", err)
		return err
	}
	submitBtn, err := g.wd.FindElement(selenium.ByCSSSelector, "#save")
	if err != nil {
		log.Errorf("没找到登录按钮 err=%v", err)
		return err
	}
	err = submitBtn.Click()
	if err != nil {
		log.Errorf("点击登录失败 err=%v", err)
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
		log.Errorf("登录管理栏出现失败 err=%v", err)
		return err
	}
	guanli, err := g.wd.FindElement(selenium.ByCSSSelector, "#Menu1_Managemen > div.item_link > a")
	if err != nil {
		log.Errorf("查找管理栏失败 err=%v", err)
		return err
	}
	err = guanli.Click()
	if err != nil {
		log.Errorf("点击管理栏失败 err=%v", err)
		return err
	}
	// 进入“设备”界面
	err = g.wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		_, err = g.wd.FindElement(selenium.ByCSSSelector, "#Menu2_Mng_Device > a")
		return err == nil, nil
	})
	if err != nil {
		log.Errorf("等待设备栏出现失败 err=%v", err)
		return err
	}
	shebei, err := g.wd.FindElement(selenium.ByCSSSelector, "#Menu2_Mng_Device > a")
	if err != nil {
		log.Errorf("查找设备栏失败 err=%v", err)
		return err
	}
	err = shebei.Click()
	if err != nil {
		log.Errorf("点击设备栏失败 err=%v", err)
		return err
	}
	// 需要切换iframe
	err = g.wd.SwitchFrame("frameContent")
	if err != nil {
		log.Errorf("切换到 id=frameContent的iframe失败 err=%v", err)
		return err
	}
	// 找到“重启”按钮
	err = g.wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		_, err = g.wd.FindElement(selenium.ByCSSSelector, "#Restart_button")
		return err == nil, nil
	})
	if err != nil {
		log.Errorf("没等到重启按钮出现 err=%v", err)
		return err
	}
	chongqi, err := g.wd.FindElement(selenium.ByCSSSelector, "#Restart_button")
	if err != nil {
		log.Errorf("查询重启按钮失败 err=%v", err)
		return err
	}
	// 点击“重启”按钮
	err = chongqi.Click()
	if err != nil {
		log.Errorf("点击重启按钮失败 err=%v", err)
		return err
	}
	// “确认”alert弹窗
	return g.wd.AcceptAlert()
}

// screenShort 将当前画面记录到本地
func (g *Job) screenShort() error {
	b, err := g.wd.Screenshot()
	if err != nil {
		log.Errorf("截图当前页面失败 err=%v", err)
		return err
	}
	return os.WriteFile("tmp.png", b, 0777)
}

// pageSource 将当前页面的源码保存到本地
func (g *Job) pageSource() error {
	b, err := g.wd.PageSource()
	if err != nil {
		log.Errorf("打印当前页面源码失败 err=%v", err)
		return err
	}
	return os.WriteFile("index.html", []byte(b), 0777)
}

// reboot 重启流程
func (g *Job) reboot() error {
	var err error
	g.wd, err = g.newChrome()
	if err != nil {
		log.Errorf("初始化 chromedriver 失败了 err=%v", err)
		return err
	}
	defer g.wd.Quit()
	defer g.pageSource()
	defer g.screenShort()
	log.Infof("start login")
	err = g.login(g.username, g.password)
	if err != nil {
		log.Errorf("login faild err=%v", err)
		return err
	}
	return g.switchToReboot()
}

func (g *Job) execJob() {
	g.ctx = context.TODO()
	err := g.reboot()
	if err != nil {
		log.Errorf("reboot faild %v", err)
	}
}

// main 主入口
func main() {
	var err error
	se := NewJob()
	_, err = se.cron.AddFunc(se.crontab, se.execJob)
	if err != nil {
		log.Errorf("计划任务创建失败 err=%v", err)
		panic(err)
	}
	se.cron.Run()
}
