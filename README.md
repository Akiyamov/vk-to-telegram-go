# Telegramrepost

## Скопировать репозиторий
```  
$ git clone https://gitlab.com/Akiyamov/telegramrepost  
$ cd telegramrepost  
```
## Запуск на своем сервере 
Для запуска требуется создать [сервисный ключ в ВК](https://dev.vk.com/ru/api/access-token/getting-started) и [бота в Telegram](https://t.me/BotFather). После этого токены передаются в контейнер через переменные. Также требуется получить ID группы в ВК, ID чата в Telegram и версию API ВК.  
Для получения ID группы можно зайти на любой пост и взять число между `wall` и `_число` включая дефис.
Последняя версия API ВК указана [по ссылке](https://dev.vk.com/ru/reference/versions)  
ID чата в Telegram можно получить через веб-версию. Нужно зайти в чат и в ссылке будет указан ID, копировать с дефисом.  
```
$ docker run  
--name repost-container  
--rm  
-e VK_TOKEN=""  
-e VK_API_VERSION=""  
-e VK_GROUP_ID=""  
-e TG_TOKEN=""  
-e TG_CHAT_ID=""  
akiyamov/telegramrepost
```