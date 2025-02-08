# backtrace

[![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fbacktrace&count_bg=%2323E01C&title_bg=%23555555&icon=sonarcloud.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://hits.seeyoufarm.com)

[![Build and Release](https://github.com/oneclickvirt/backtrace/actions/workflows/main.yaml/badge.svg)](https://github.com/oneclickvirt/backtrace/actions/workflows/main.yaml)

[<img src="https://api.gitsponsors.com/api/badge/img?id=794459615" height="20">](https://api.gitsponsors.com/api/badge/link?p=qZ+xPDJJLB3lt/bXVIj9S/1wpaPBClR9U4f95dSfrDpGr+ftyXD1Y4lHMWabR9FVAzY/qBqT2SmnV4RA0zPKncMfplAZxsVoaht4dis+6CekCtdkoBywGbKhDcP4SCrCoPKO+bIXD73oxN3BfKfwDw==)

三网回程路由线路测试

基于 https://github.com/zhanghanyun/backtrace 的重构和优化，与原版存在很大不同

路由的线路判断最终还是得人工判断的才准确，本项目测试结果仅供参考

## 功能

- [x] 检测回程显示IPV4地址时的线路(使用1500字节的包)，不显示IP地址时显示ASN检测不到，原版[backtrace](https://github.com/zhanghanyun/backtrace)也支持
- [x] 支持对```4837```、```9929```和```163```线路的判断，原版[backtrace](https://github.com/zhanghanyun/backtrace)也支持
- [x] 支持对```CN2GT```和```CN2GIA```线路的判断，原版[backtrace](https://github.com/zhanghanyun/backtrace)不支持，原版全部识别为```CN2```了
- [x] 支持对```CMIN2```和```CMI```线路的判断，原版[backtrace](https://github.com/zhanghanyun/backtrace)也支持，但所支持的IP区间不一样，本项目更多
- [x] 支持对整个回程路由进行线路分析，与原版[backtrace](https://github.com/zhanghanyun/backtrace)仅进行一次判断不同
- [x] 修复原版[backtrace](https://github.com/zhanghanyun/backtrace)对IPV4地址信息获取时json解析失败依然打印信息的问题，本项目忽略错误继续执行路由线路查询
- [x] 增加对全平台的编译支持，原版[backtrace](https://github.com/zhanghanyun/backtrace)仅支持linux平台的amd64和arm64架构

## TODO

- [ ] 增加对CTG回程的判断
- [ ] 使用nexttrace进行路由检测，备用方案才使用本地路由检测
- [ ] 自动检测汇聚层，裁剪结果不输出汇聚层后的线路

## 使用

下载、安装、更新

```shell
curl https://raw.githubusercontent.com/oneclickvirt/backtrace/main/backtrace_install.sh -sSf | bash
```

或

```
curl https://cdn.spiritlhl.net/https://raw.githubusercontent.com/oneclickvirt/backtrace/main/backtrace_install.sh -sSf | bash
```

使用

```
backtrace
```

或

```
./backtrace
```

进行测试

无环境依赖，理论上适配所有系统和主流架构，更多架构请查看 https://github.com/oneclickvirt/backtrace/releases/tag/output

```
Usage: backtrace [options]
  -e    Enable logging
  -h    Show help information
  -s    Disabe show ip info (default true)
  -v    Show version
```

## 卸载

```
rm -rf /root/backtrace
rm -rf /usr/bin/backtrace
```

## 在Golang中使用

```
go get github.com/oneclickvirt/backtrace@latest
```

## 概览图

![图片](https://github.com/oneclickvirt/backtrace/assets/103393591/4688f99f-0f02-486f-8ffc-78d30f2c2f95)

![图片](https://github.com/oneclickvirt/backtrace/assets/103393591/2812a47d-4e6b-4091-9bb9-596af6c3c8bc)

![图片](https://github.com/oneclickvirt/backtrace/assets/103393591/2e5cc625-e0da-41ff-85ff-9d21c01114a3)
