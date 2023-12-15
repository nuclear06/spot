- [spot](#spot)
  - [特性](#特性)
  - [配置](#配置)
  - [使用](#使用)
    - [二进制](#二进制)
    - [从源码编译](#从源码编译)
    - [systemd (service)](#systemd-service)
    - [docker](#docker)
    - [可视化(optional)](#可视化optional)
  - [鸣谢](#鸣谢)
  - [声明](#声明)

## spot

一个低交互的 SSH 蜜罐，基于 golang 的 crypto/ssh 库开发。

### 特性

- 支持未验证连接、密码登录、秘钥登录三种方式。
- 可记录远程 ip，用户密码、私钥指纹。
- 可生成一个简易终端接受远程命令并记录。
- 具有默认配置，开箱即用；也可使用 yml 进一步配置。
- 日志可以 json 格式输出到文件，支持日志文件滚动、gzip 压缩。

### 配置

> 使用`spot -h`和`spot conf -h`可输出完整帮助信息 \
> 可使用`spot conf -i`生成默认配置文件,`spot conf -c <config>`来打印运行时配置

程序默认在当前路径寻找 config.yml 配置文件，不存在会自动退出。可使用`spot conf -i`生成默认配置文件。

```shell
#  最简使用
./spot conf -i
./spot
```

完整的配置文件及说明(默认配置): \
[中文](config-zh-example.yml) | [English](config-en-example.yml)

### 使用

#### 二进制

从 Github [release](https://github.com/nuclear06/spot/releases) 下载使用即可

#### 从源码编译

本程序在 Golang1.21.4 下开发，不保证其他版本一定可行

```shell
git clone https://github.com/nuclear06/spot.git
make
```

#### systemd (service)

```shell
[Unit]
Description=SSH honeypot
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/path/to/spot #-d /path/to/config.yml
Restart=always

[Install]
WantedBy=multi-user.target
```

#### docker

使用 distroless/debian11 作为基础镜像，保证镜像足够小(8.56MB)且安全。

```shell
# 开箱即用
docker run --rm -p 22:2023 \
    --name spot \
    nuclear06/spot:latest

# 映射配置文件和日志文件
docker run -d -p 22:2023 \
    -v ./config.yml:/spot/config.yml \
    -v ./logs:/spot/logs \
    --restart=always \
    --name spot \
    nuclear06/spot:latest
```

#### 可视化(optional)

项目提供简单的可视化 Python 脚本，详情请查看[toos.py](./script/visualize/tools.py)\
使用示例见[Jupyter Notebook](./script/visualize/example.ipynb) \
效果示例[example1](./assert/img/example1.png) | [example2](./assert/img/example2.png)

> **/ ! \\** 注意,IP Map使用了`IP2Location LITE`(Free)数据库,请自行前往[官网](https://lite.ip2location.com/databaseip-country-region-city-latitude-longitude)下载

### 鸣谢

感谢以下项目,为本项目提供了参考：\
[sshesame](https://github.com/jaksi/sshesame) \
[go-sshoney](https://github.com/ashmckenzie/go-sshoney) \
[go-ssh-examples](https://github.com/Scalingo/go-ssh-examples)

### 声明

> This site or product includes IP2Location LITE data available from <a href="https://lite.ip2location.com">https://lite.ip2location.com</a>.
