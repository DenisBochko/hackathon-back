# Заготовка Backend для хакатона осень 2025

## Запуск

### Запуск для разработчика (конфиг по дефолту [config.template.yml](config/config.template.yml))
```bash
# Запускаем зависимые сервисы
docker compose -f docker-compose.dev.yml up -d

# Генерируем .pem ключи
task keygen

# Запуск приложения
task run 
```

### Запуск для фронтендера (конфиг по дефолту [config.docker.yml](config/config.docker.yml))
```bash
# Клонируем репозиторий 
git clone https://github.com/DenisBochko/hackathon-back.git

# Переходим в директорию проекта
cd hackathon-back

# Запускаем в Docker и всё
docker compose up -d --build 

# Можно потыкать конфиги, но это не рекомендуется 
```

## Использование при локальном запуске

- Документация находится по адресу: `http://localhost:8080/api/docs/swagger/index.html`
- Тестовый smtp сервер: `http://localhost:8025`
- Базовый путь `/api`
- Дефолтный manager: "email": "manager@gmail.com", "password": "12345678"

## Общие рекомендации по разработке

1. Внимательно прочитай Taskfile. Используй `task`, чтобы посмотреть доступные команды.
2. Для создания новой миграции используй task migrate-create.
3. Форматируй проект.
4. Запускай почаще линтеры и исправляй ошибки, ими найденные.
5. Если обновил что-то в swagger документации, то вызывай task swagger, чтобы подтянуть изменения
