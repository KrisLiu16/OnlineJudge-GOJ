#!/bin/bash

# 检查 Docker 是否已安装
if command -v docker &> /dev/null
then
    echo "Docker 已安装，跳过安装步骤。"
else
    echo "Docker 未安装，正在安装 Docker..."

    # 更新包索引
    sudo apt-get update

    # 安装必要的包
    sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common

    # 添加 Docker 的官方 GPG 密钥
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

    # 添加 Docker 的稳定版仓库
    sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

    # 更新包索引
    sudo apt-get update

    # 安装 Docker
    sudo apt-get install -y docker-ce

    # 启动 Docker 服务
    sudo systemctl start docker
    sudo systemctl enable docker

    # 验证 Docker 是否安装成功
    docker --version
fi

# 检查 Docker Compose 是否已安装
if command -v docker-compose &> /dev/null
then
    echo "Docker Compose 已安装，跳过安装步骤。"
else
    echo "Docker Compose 未安装，正在安装 Docker Compose..."

    # 下载 Docker Compose
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

    # 赋予执行权限
    sudo chmod +x /usr/local/bin/docker-compose

    # 验证 Docker Compose 是否安装成功
    docker-compose --version
fi

# 使用 docker-compose.yml 拉取镜像
echo "正在使用 docker-compose 拉取镜像..."
sudo docker-compose -f docker/docker-compose.yml up -d