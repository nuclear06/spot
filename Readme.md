## spot
一个低交互的SSH蜜罐，基于golang的crypto/ssh库开发。

### 特性
+ 支持未验证连接、密码登录、秘钥登录三种方式。
+ 可记录远程ip，用户密码、私钥指纹。
+ 可生成一个简易终端接受远程命令并记录。
+ 具有默认配置，开箱即用；也可使用yml进一步配置。
+ 日志可以json格式输出到文件，支持日志文件滚动、gzip压缩。

### 配置
> 使用`spot -h`和`spot conf -h`可输出完整帮助信息  \
> 可使用`spot conf -i`生成默认配置文件,`spot conf -c <config>`来打印运行时配置

程序默认在当前路径寻找config.yml配置文件，不存在会自动退出。可使用`spot conf -i`生成默认配置文件。

 ```shell
 #  最简使用
 ./spot conf -i
 ./spot
 ```

完整的配置文件及说明(默认配置): \
[中文](config-zh-example.yml) | [English](config-en-example.yml)

### 安装

#### 二进制

[release]()

#### 从源码编译

本程序在Golang1.21.4下开发，不保证其他版本一定可行

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
使用distroless/debian11作为基础镜像，保证镜像足够小(8.56MB)且安全。
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

### 参考

感谢以下项目提供的帮助：\
[sshesame](https://github.com/jaksi/sshesame) \
[go-sshoney](https://github.com/ashmckenzie/go-sshoney) \
[go-ssh-examples](https://github.com/Scalingo/go-ssh-examples)