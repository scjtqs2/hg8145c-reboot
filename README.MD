# 前言
> 过年家里没人，光猫死了之后，直接断网。太难受了。电信给的光猫来垃圾，也是在没办法了。随之产生了定时重启它的想法。
> 
> 网上找了一圈，都是要破解超管账号然后解锁telnet的。我这长期在外，只能远程操作，没办法断网断电折腾，也就没办法拿到超管密码了，用自带的用户账号进去看了下，有重启按钮。😁
> 
> 简单抓了下api。。好吧还是带表单验证的，算了，不深入折腾。直接上无头浏览器吧。

# 说明
> 基于selenium实现

这个型号的直接用。其他型号的，得自己抓页面标签了。

docker-compose.yaml里面的镜像支持amd64和arm64。

## docker-coompose
```yaml
services:
  # 基于chrome的无头浏览器
  chrome:
    image: seleniarm/standalone-chromium:124.0-20240427
    restart: always
    expose:
      - 4444
    shm_size: 2gb
  job:
    image: ghcr.io/scjtqs2/hg8145c-reboot:edge
    pull_policy: always              # 每次都更新
    restart: always
    environment:
      - TZ=Asia/Shanghai
      - LANG=zh_CN.UTF-8
      - SELENIUM_ADDR="http://chrome:4444"
      - LOGIN_URL="http://192.168.1.1" # 这里填你的光猫web地址
      - LOGIN_USERNAME=useradmin       # 光猫贴纸上的用户名
      - LOGIN_PASSWORD=password        # 光猫贴纸上的密码
      - CRONTAB="0 0 * * *"            # 定时重启时间，默认每天凌晨0点重启。规程参考linux的crontab

```