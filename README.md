# GO! Judge
[![Go Reference](https://pkg.go.dev/badge/github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend.svg)](https://pkg.go.dev/github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend) [![Go Report Card](https://goreportcard.com/badge/github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend)](https://goreportcard.com/report/github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend) [![Release](https://img.shields.io/github/v/tag/KrisLiu16/OnlineJudge-GOJ)](https://github.com/KrisLiu16/OnlineJudge-GOJ/releases/latest) ![Go CI Status](https://github.com/KrisLiu16/OnlineJudge-GOJ/actions/workflows/go.yml/badge.svg)

[![Vue](https://img.shields.io/badge/vue-3.3.4-brightgreen.svg?style=flat-square)](https://vuejs.org/)
[![Go](https://img.shields.io/badge/go-1.22-blue.svg?style=flat-square)](https://golang.org/)
[![Gin](https://img.shields.io/badge/gin-1.9.1-blue.svg?style=flat-square)](https://gin-gonic.com/)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/krisliu16/onlinejudge-goj)

> #### 基于 Go 和 Vue3 的计算机程序在线评测系统

[English Document](README-EN.md)

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
   docker-compose up -d --build
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
- [X] WebSocket 实时反馈
- [ ] 其它问题修复...

## 贡献

欢迎提交 Issue 和 Pull Request。

## 许可

[MIT License](LICENSE)

## 项目地址

[GitHub 项目地址](https://github.com/KrisLiu16/OnlineJudge-GOJ)
