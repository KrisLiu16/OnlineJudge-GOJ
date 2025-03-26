# GO! Judge

[![Vue](https://img.shields.io/badge/vue-3.3.4-brightgreen.svg?style=flat-square)](https://vuejs.org/)
[![Go](https://img.shields.io/badge/go-1.22-blue.svg?style=flat-square)](https://golang.org/)
[![Gin](https://img.shields.io/badge/gin-1.9.1-blue.svg?style=flat-square)](https://gin-gonic.com/)

> #### 基于 Go 和 Vue3 的计算机程序在线评测系统

[English Document](README-EN.md)

## 概览

- 基于 Docker，一键部署
- 前后端分离，微服务架构
- 支持多种编程语言
- 实时评测，快速反馈
- 渐进式并且统一的UI风格
- 完善的权限管理系统
- 支持 Markdown & LaTeX

![image](https://github.com/user-attachments/assets/91321577-8c96-4c9e-b496-231ede1ecf2f)

![image](https://github.com/user-attachments/assets/5e30fd9e-196f-497e-b13c-42b9277774c3)

![image](https://github.com/user-attachments/assets/0655efb1-b5c1-4d51-a18a-f98380ea0711)

![image](https://github.com/user-attachments/assets/6582aa24-10f3-49dc-9b5c-df1be448ab09)

![image](https://github.com/user-attachments/assets/42c2090e-fbd2-4b9b-9e4d-cb4752ed166d)

![image](https://github.com/user-attachments/assets/3a2c8a48-ac77-4ea0-a19b-60bf5b6a2615)

![image](https://github.com/user-attachments/assets/1fe70f65-e516-4437-b347-7776de206851)


主要模块：

- 前端 (Vue3): [goj-frontend](https://github.com/KrisLiu16/OnlineJudge-GOJ/tree/main/goj-frontend)
- 后端 (Golang): [goj-backend](https://github.com/KrisLiu16/OnlineJudge-GOJ/tree/main/goj-backend)
- 判题服务 (go-judge): [go-judge](https://github.com/criyle/go-judge)

## 安装

### 提示：如果不会使用 Docker

如果你不熟悉 Docker 的使用，可以使用以下脚本快速安装 Docker 和本项目（如果国内用户很慢可能需要使用代理）：

#### 安装 Docker

在项目根目录下，你可以使用以下命令来克隆项目并运行安装脚本：

1. 克隆项目源代码：

   ```bash
   git clone https://github.com/KrisLiu16/OnlineJudge-GOJ.git
   cd OnlineJudge-GOJ
   ```

2. 运行安装脚本：

   ```bash
   sudo bash install.sh
   ```

#### 方法一：使用 Docker 拉取镜像

1. 确保你已经安装了 Docker 和 Docker Compose。
2. 直接从 GitHub 拉取 `docker-compose.yml` 文件：

   ```bash
   curl -O https://raw.githubusercontent.com/KrisLiu16/OnlineJudge-GOJ/main/docker/docker-compose.yml
   ```

3. 在终端中运行以下命令以启动服务：

   ```bash
   docker-compose up -d
   ```

这种方法适合懒人或不会安装 Docker 的用户。

### 方法二：从源代码构建

1. 克隆项目源代码：

   ```bash
   git clone https://github.com/KrisLiu16/OnlineJudge-GOJ.git
   cd OnlineJudge-GOJ
   ```

2. 在项目根目录下，使用 `docker-compose.yml` 文件构建前后端：

   ```bash
   docker compose up -d --build
   ```

3. 如果需要修改源代码，可以在本地编辑，然后重新构建。

## 特性

### 现代化的用户界面
- 响应式设计，支持移动端
- 多种色彩模式支持
- 简洁直观的操作体验，现代化的UI，美观

### 强大的评测系统
- 支持多种编程语言
- 实时评测反馈
- 安全的沙箱环境

### 完善的比赛系统
- 支持多种比赛模式
- 实时排名更新

### 社区功能
- 题解讨论
- 用户排名
- 个人主页

## 技术栈

### 前端
- Vue 3
- TypeScript
- Vite
- Pinia
- Vue Router

### 后端
- Golang
- Gin
- GORM
- Redis
- MySQL

### 部署
- Docker
- Docker Compose
- Nginx

## 开发计划

- [X] SPJ完全支持
- [ ] WebSocket 实时反馈
- [ ] ...

## 贡献

欢迎提交 Issue 和 Pull Request。

## 许可

[MIT License](LICENSE)

## 项目地址

[GitHub 项目地址](https://github.com/KrisLiu16/OnlineJudge-GOJ)
