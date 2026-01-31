# Order Information Service

Демонстрационный микросервис для приёма, хранения и отображения информации о заказах.  

---

## Описание

Сервис получает события о заказах из **Kafka**, сохраняет их в **PostgreSQL**, кэширует в памяти и предоставляет доступ к данным через **HTTP API** и простой **Web UI**.  
Дополнительно реализован мониторинг с помощью **Prometheus** и **Grafana**.

---

## Архитектура


- Kafka — источник событий о заказах
- PostgreSQL — основное хранилище
- In-memory LRU cache — ускорение чтения
- HTTP API — получение заказа по `order_uid`
- Prometheus + Grafana — метрики и мониторинг

---

## Стек технологий

- Go 1.25
- Kafka
- PostgreSQL
- Docker / Docker Compose
- Prometheus
- Grafana
- Uber Zap (логирование)
