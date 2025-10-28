# GO! Judge

[![Go Reference](https://pkg.go.dev/badge/github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend.svg)](https://pkg.go.dev/github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend) [![Go Report Card](https://goreportcard.com/badge/github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend)](https://goreportcard.com/report/github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend) [![Release](https://img.shields.io/github/v/tag/KrisLiu16/OnlineJudge-GOJ)](https://github.com/KrisLiu16/OnlineJudge-GOJ/releases/latest) ![Go CI Status](https://github.com/KrisLiu16/OnlineJudge-GOJ/actions/workflows/go.yml/badge.svg)

[![Vue](https://img.shields.io/badge/vue-3.3.4-brightgreen.svg?style=flat-square)](https://vuejs.org/)
[![Go](https://img.shields.io/badge/go-1.22-blue.svg?style=flat-square)](https://golang.org/)
[![Gin](https://img.shields.io/badge/gin-1.9.1-blue.svg?style=flat-square)](https://gin-gonic.com/)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/krisliu16/onlinejudge-goj)

> #### A modern online judging system based on Go and Vue3

[中文文档](README.md)

Main modules:

- Frontend (Vue3): [goj-frontend](https://github.com/KrisLiu16/OnlineJudge-GOJ/tree/main/goj-frontend)
- Backend (Golang): [goj-backend](https://github.com/KrisLiu16/OnlineJudge-GOJ/tree/main/goj-backend)
- Judging service (go-judge): [go-judge](https://github.com/criyle/go-judge)

## Installation

### Tip: If you are not familiar with Docker

If you are not familiar with using Docker, you can use the following script to quickly install Docker and this project (users in China may need to use a proxy if the download is slow):

#### Install Docker

In the project root directory, you can use the following commands to clone the project and run the installation script:

1. Clone the project source code:

   ```bash
   git clone https://github.com/KrisLiu16/OnlineJudge-GOJ.git
   cd OnlineJudge-GOJ
   ```

2. Run the installation script:

   ```bash
   sudo bash install.sh
   ```

#### Method 1: Use Docker to Pull Images

1. Ensure that you have Docker and Docker Compose installed.
2. Directly pull the `docker-compose.yml` file from GitHub:

   ```bash
   curl -O https://raw.githubusercontent.com/KrisLiu16/OnlineJudge-GOJ/main/docker/docker-compose.yml
   ```

3. Run the following command in the terminal to start the service:

   ```bash
   docker-compose up -d
   ```

This method is suitable for those who prefer a quick setup or are not familiar with installing Docker.

### Method 2: Build from Source Code

1. Clone the project source code:

   ```bash
   git clone https://github.com/KrisLiu16/OnlineJudge-GOJ.git
   cd OnlineJudge-GOJ
   ```

2. In the project root directory, use the `docker-compose.yml` file to build the frontend and backend:

   ```bash
   docker-compose up -d --build
   ```

3. If you need to modify the source code, you can edit it locally and then rebuild.

## Features

### Modern User Interface
- Responsive design, mobile-friendly
- Support for multiple color modes
- Simple and intuitive user experience with a modern UI

### Powerful Evaluation System
- Supports multiple programming languages
- Real-time evaluation feedback
- Secure sandbox environment

### Comprehensive Competition System
- Supports various competition modes
- Real-time ranking updates

### Community Features
- Problem discussion
- User rankings
- Personal homepage

## Tech Stack

### Frontend
- Vue 3
- TypeScript
- Vite
- Pinia
- Vue Router

### Backend
- Golang
- Gin
- GORM
- Redis
- MySQL

### Deployment
- Docker
- Docker Compose
- Nginx

## Development Plan

- [ ] Custom themes for the frontend
- [ ] WebSocket real-time feedback
- [ ] ...

## Contribution

Contributions are welcome! Please submit Issues and Pull Requests.

## License

[MIT License](LICENSE)

## Project Address

[GitHub Project Address](https://github.com/KrisLiu16/OnlineJudge-GOJ)
