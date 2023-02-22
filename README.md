# php-detect

根据当前项目配置自动选择对应版本的PHP可执行文件

# 安装

1. 将编译后的文件放到系统环境变量
2. 将原先的全局PHP删除
3. 在配置文件中指定各个版本路径

# 使用

## 预定义各个版本所在路径

在`家`目录新建`php-detect`文件内容如下

```text
default=8.0
8.0=/usr/local/bin/php80
8.2=/usr/local/bin/php82
```

## 环境变量传入

> Linux & MacOS

指定以8.0版本php执行
`PV=8.0 php ....`

## phpver文件指定

在当前项目中新建phpver文件，文件内容为版本号

`php ....`

## 自动识别当前项目的composer.json中的php版本

`php ....`