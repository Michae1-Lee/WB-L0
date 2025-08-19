# WB Order Service

## 📌 Описание
WB Order Service - это backend-сервис для работы с заказами.  
Сервис принимает данные заказов из Kafka, сохраняет их в PostgreSQL и во внутренний кэш на основе `map`,  
а также предоставляет REST API для получения информации о заказах.

---

## 🚀 Функциональность
- Приём заказов из **Kafka (consumer)**
- Генерация тестовых заказов и отправка в **Kafka (producer)**
- Сохранение заказов в **PostgreSQL** и **in-memory cache**
- REST API для получения заказа по `order_uid`

---

## 🛠️ Технологии
- **Go**
- **PostgreSQL**
- **Kafka**
- **In-memory cache (map)**
- **Zap logger**
- **Docker + Docker Compose**

---

## 📡 API
`GET /order/{order_uid}` - получить заказ по уникальному идентификатору  

---

## 📡 Запуск
`make run` - Поднимет PostgreSQL, Kafka, Приложение

`make down` - Остановка
