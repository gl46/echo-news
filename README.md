# EchoNews

> 🌐 A bilingual news reader powered by BBC RSS — read English news with Chinese translations side by side.
>
> 基于 BBC RSS 的中英双语新闻阅读器，支持 LLM 驱动的实时翻译。

## ✨ Features

- 📡 自动抓取 BBC RSS 新闻并入库，支持去重
- 📖 中英文对照阅读
- 

## 🛠 Tech Stack

| Layer    | Tech                    |
| -------- | ----------------------- |
| Backend  | Go                      |
| Frontend | Vue 3 + Vite + Router   |
| Database | SQLite                  |
| Translation | LLM API              |

## 🚀 Getting Started

### Prerequisites

- Go 



## 📐 API Overview

| Method | Endpoint               | Description          |
| ------ | ---------------------- | -------------------- |
| GET    | `/api/articles`        | 获取文章列表         |
| GET    | `/api/articles/:id`    | 获取文章详情         |
| POST   | `/api/translate/:id`   | 触发翻译并返回结果   |

## 🗺 Roadmap

- [x] 确定项目名与技术方案
- [ ] **MVP**：BBC RSS 抓取 → 文章列表 → 详情页 → 手动翻译 → 跑通演示
- [ ] 多 RSS 源 / 分类过滤 / 翻译缓存 / 定时抓取
- [ ] TTS 播报 / LLM 摘要 / 关键词提取

## 📄 License

[MIT](LICENSE)
