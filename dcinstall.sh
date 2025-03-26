
#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# 检查是否以 root 运行
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}错误：请使用 sudo 或以 root 用户运行此脚本！${NC}"
    exit 1
fi

# 安装 Docker
install_docker() {
    echo -e "${YELLOW}[1/4] 正在安装 Docker...${NC}"

    # 卸载旧版本
    echo -e "${YELLOW}>>> 卸载旧版本 Docker...${NC}"
    apt remove -y docker docker-engine docker.io containerd runc 2>/dev/null

    # 安装依赖
    echo -e "${YELLOW}>>> 安装依赖...${NC}"
    apt update -y && apt install -y ca-certificates curl gnupg lsb-release

    # 添加 Docker GPG 密钥
    echo -e "${YELLOW}>>> 添加 Docker GPG 密钥...${NC}"
    mkdir -p /etc/apt/keyrings && \
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

    if [ $? -ne 0 ]; then
        echo -e "${RED}错误：无法添加 Docker GPG 密钥！${NC}"
        exit 1
    fi

    # 添加 Docker 仓库
    echo -e "${YELLOW}>>> 添加 Docker 仓库...${NC}"
    echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
    $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

    # 安装 Docker Engine
    echo -e "${YELLOW}>>> 安装 Docker Engine...${NC}"
    apt update -y && apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

    if [ $? -ne 0 ]; then
        echo -e "${RED}错误：Docker 安装失败！${NC}"
        exit 1
    fi

    # 启动 Docker 服务
    echo -e "${YELLOW}>>> 启动 Docker 服务...${NC}"
    systemctl enable --now docker

    # 测试 Docker
    echo -e "${YELLOW}>>> 测试 Docker 是否安装成功...${NC}"
    docker run hello-world >/dev/null 2>&1

    if [ $? -ne 0 ]; then
        echo -e "${RED}错误：Docker 测试失败！${NC}"
        exit 1
    fi

    echo -e "${GREEN}Docker 安装成功！${NC}"
}

# 安装 Docker Compose
install_docker_compose() {
    echo -e "${YELLOW}[2/4] 正在安装 Docker Compose...${NC}"

    # 检查是否已安装 docker-compose-plugin（推荐）
    if command -v docker compose &>/dev/null; then
        echo -e "${GREEN}>>> Docker Compose (docker compose) 已安装！${NC}"
        return
    fi

    # 如果没有，安装旧版 docker-compose
    echo -e "${YELLOW}>>> 安装旧版 docker-compose...${NC}"
    apt install -y docker-compose

    if [ $? -ne 0 ]; then
        echo -e "${RED}错误：Docker Compose 安装失败！${NC}"
        exit 1
    fi

    echo -e "${GREEN}Docker Compose 安装成功！${NC}"
}

# 配置当前用户免 sudo 使用 Docker
configure_user() {
    echo -e "${YELLOW}[3/4] 配置当前用户免 sudo 使用 Docker...${NC}"

    CURRENT_USER=$(logname)
    if [ -z "$CURRENT_USER" ]; then
        CURRENT_USER=$(whoami)
    fi

    echo -e "${YELLOW}>>> 将用户 ${CURRENT_USER} 加入 docker 组...${NC}"
    usermod -aG docker "$CURRENT_USER"

    if [ $? -ne 0 ]; then
        echo -e "${RED}错误：无法将用户加入 docker 组！${NC}"
        exit 1
    fi

    echo -e "${GREEN}配置完成！请重新登录或运行 'newgrp docker' 生效。${NC}"
}

# 显示安装结果
show_result() {
    echo -e "${GREEN}[4/4] 安装完成！${NC}"
    echo -e "${YELLOW}>>> Docker 版本：${NC}"
    docker --version
    echo -e "${YELLOW}>>> Docker Compose 版本：${NC}"
    if command -v docker compose &>/dev/null; then
        docker compose version
    else
        docker-compose --version
    fi
    echo -e "${GREEN}✅ 所有步骤已完成！${NC}"
}

# 主函数
main() {
    install_docker
    install_docker_compose
    configure_user
    show_result
}

main