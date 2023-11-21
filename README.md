# Telegramrepost

Телеграм-бот для репостов из групп ВК в Telegram, образ которого весит менее 10 мегабайт и практически не потребляет ресурсы.

## Используемые модули

- [vksdk](https://github.com/SevereCloud/vksdk) - Модуль для упрощения работы с API ВК.
- [godotenv](https://github.com/joho/godotenv) - Модуль для загрузки переменных среды из файла.

## Запуск на своем сервере

Для запуска бота требуется получить креды для ВК и Telegram, как это сделать описано [здесь](https://gitlab.com/Akiyamov/telegramrepost/-/wikis/%D0%9F%D0%BE%D0%BB%D1%83%D1%87%D0%B5%D0%BD%D0%B8%D0%B5-%D0%B4%D0%B0%D0%BD%D0%BD%D1%8B%D1%85-%D0%B4%D0%BB%D1%8F-%D0%B7%D0%B0%D0%BF%D1%83%D1%81%D0%BA%D0%B0-%D0%B1%D0%BE%D1%82%D0%B0).  
Также для запуска требуется локальное API Telegram, так как это позволяет увеличить размер загружаемых файлов. Самым простым способом, для этого запустите контейнер aiogram/telegram-bot-api следующим образом:  
```bash
docker run -d -p 8081:8081 /
--name=telegram-bot-api /
--restart=always /
-v telegram-bot-api-data:/var/lib/telegram-bot-api /
-e TELEGRAM_API_ID=<api_id> /
-e TELEGRAM_API_HASH=<api-hash> /
aiogram/telegram-bot-api:latest
```  
Вместо `api_id` и `api-hash` подставьте свои креды.  
Для запуска требуются две директории: `video` и любая, в которой будет находиться файл `.env`. Первая используется для временного хранения файлов, вторая для хранения файла с переменными среды. Желательно создать отдельные директории с правами доступа только для пользователя, который будет использоваться внутри контейнера.

### Запуск без контейнера  

Для запуска бота нужно установить [Go](https://go.dev/).  
После этого нужно скопировать репозиторий себе на ПК.
```bash
$ git clone https://gitlab.com/Akiyamov/telegramrepost  
$ cd telegramrepost  
```
После того как репозиторий и Go установлен можно скачать mod.
```bash
$ go mod download
```
Удалите функцию `LoadEnv()` и ее вызов в функции `main()` и замените следующие значения на свои:
```golang
vk_access_token = os.Getenv("VK_TOKEN")
vk_api_version = os.Getenv("VK_API_VERSION")
vk_owner_id = os.Getenv("VK_GROUP_ID")
telegram_bot_token = os.Getenv("TG_TOKEN")
telegram_chat_id = os.Getenv("TG_CHAT_ID")
telegram_temp_chat_id = os.Getenv("TG_TEMP_CHAT_ID")
```
После этого можно запускать приложение:
```bash
$ go run main.go
```

### Запуск в контейнере  

После каждого пуша в главную ветку репозитория CI/CD собирает новый образ контейнера, поэтому последнюю версию можно скачать следующим образом:
```bash
$ docker pull akiyamov/telegramrepost
```  
После этого контейнер можно запустить:
```bash
docker run -u uid:gid \
--restart on-failure \
--name repost-container -d \
--network host -v /dest/to/fold:/video \
-v /dest/to/fold/.env:/opt/.env \
akiyamov/telegramrepost:latest
```
`uid`, `gid`, а также пути к директориям нужно указать свои.