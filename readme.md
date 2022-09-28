# Описание

## Система управления кодовыми строками "Мастер ключей"

---

### Общие требования

Система представляет собой HTTP API со следующей бизнес-логикой:

* регистрация, аутентификация и авторизация пользователей;
* приём ключа для доступа к кодовой строке и выдача этой строки авторизированному пользователю;
* учёт и ведение списка сгенерированных кодовых строк и ключей к ним;
* отслеживание использования кодовых строк зарегистрированными пользователями;
* блокировка доступа к кодовому слову согласно на основании количества использований и срока действия кодового слова.

### Абстрактная схема взаимодействия с системой

Ниже представлена абстрактная бизнес-логика взаимодействия пользователя с системой:

1. Новый пользователь регистрируется в системе «Мастер ключей"».
2. Существующий пользователь входит в систему «Мастер ключей"».
3. Пользователь запрашивает ключ для получения кодовой строки.
4. Система генерирует кодовую строку и ключ для доступа для авторизованного пользователя.
5. Пользователь использует ключ для получения кодовой строки.
6. Пользователь может обновлять токены после истечения TTL Access токена без повторного входа.

### Общие ограничения и требования
#### Хранение
* хранилище данных — PostgreSQL (интерфейс для map объявлен, но не реализован)
#### HTTP
* возвращаемые коды ответа сервера - на усмотрение кандидата
* сервер не использует сжатие данных
* маршрутизатор [gin](https://github.com/gin-gonic/gin)
* управление токенами на базе [jwt](https://github.com/dgrijalva/jwt-go)
#### Ключ и Кодовая строка
* ключ
    * любая длина (выбрана длина 10 символов - буквы и цифры)
    * любые символы
* кодовая строка
    * длина 500 символов - генерируется на базе [pwgen](https://github.com/chr4/pwgen)
    * любые символы (буквы, цифры и символы)
* ключи и кодовые строки уникальны в пределах системы
* время действия
    * ключ (и кодовая строка) действительны в течение 72 часов
    * кодовую строку по ключу можно получить не более 3 раз
* ключ может использовать только авторизированный пользователь, для которого был сгенерирован данный ключ
* в один момент времени у пользователя может быть доступен только одна пара ключ-строка
* при запросе нового ключа генерируется новая пара ключ-строка. а все существующие становятся недоступны

### Сводное HTTP API

Система управления кодовыми строками "Мастер ключей" предоставляет следующие HTTP-хендлеры:

* `POST /api/user/register` — регистрация пользователя;
* `POST /api/user/login` — аутентификация пользователя;
* `POST /api/user/refresh` - обновление токенов по Refresh токену;
* `GET /api/secret` — получение ключа для получения кодовой строки;
* `GET /api/secret/{ключ}` — получение кодовой строки.

#### Регистрация пользователя

Хендлер: `POST /api/user/register`.

Регистрация производится по паре логин/пароль. Каждый логин должен быть уникальным.
После успешной регистрации должна происходить автоматическая аутентификация пользователя.

Формат запроса:

```
POST /api/user/register HTTP/1.1
Content-Type: application/json
...

{
	"login": "<login>",
	"password": "<password>"
}
```

Возможные коды ответа:

- `200` — пользователь успешно зарегистрирован и аутентифицирован;
  Формат ответа:

    ```
    200 OK HTTP/1.1
    Content-Type: application/json
    Authorization: Bearer <access_token>
    ...

    {
        "accessToken": "<access_token>",
        "refreshToken": "<refresh_token>>",
    }
    ```
- `400` — неверный формат запроса;
- `409` — логин уже занят;
- `500` — внутренняя ошибка сервера.

#### Аутентификация пользователя

Хендлер: `POST /api/user/login`.

Аутентификация производится по паре логин/пароль.

Формат запроса:

```
POST /api/user/login HTTP/1.1
Content-Type: application/json
...

{
	"login": "<login>",
	"password": "<password>"
}
```

Возможные коды ответа:

- `200` — пользователь успешно аутентифицирован;
  Формат ответа:

    ```
    200 OK HTTP/1.1
    Content-Type: application/json
    Authorization: Bearer <access_token>
    ...

    {
        "accessToken": "<access_token>",
        "refreshToken": "<refresh_token>>",
    }
    ```
- `400` — неверный формат запроса;
- `401` — неверная пара логин/пароль;
- `500` — внутренняя ошибка сервера.

#### Обновление токенов

Хендлер: `POST /api/user/refresh`.

Хендлер доступен всем пользователям.

На стороне пользователя должна быть реализована система хранения Refresh токена.

Формат запроса:

```
POST /api/user/refresh HTTP/1.1
Content-Type: application/json
...

{
    "token" : "<refresh_token>"
}
```

Возможные коды ответа:

- `200` — новая пара токенов успешно предоставлена;
  Формат ответа:

    ```
    200 OK HTTP/1.1
    Content-Type: application/json
    Authorization: Bearer <access_token>
    ...

    {
        "AccessToken": "<access_token>",
        "RefreshToken": "<refresh_token>>",
    }
    ```
- `401` — токен не предоставлен, либо не прошел проверку;
- `500` — внутренняя ошибка сервера.

#### Генерация ключа и кодовой строки

Хендлер: `GET /api/user/secret`.

Хендлер доступен только авторизованному пользователю.

Формат запроса:

```
GET /api/user/secret HTTP/1.1
Content-Length: 0
```

Возможные коды ответа:

- `200` — успешная обработка запроса.

  Формат ответа:

    ```
    200 OK HTTP/1.1
    Content-Type: application/json
    Authorization: ...
    ...
    
    {
        "key": "<key>"
    }
    ```
- `401` — пользователь не авторизован.
- `402` — неверный формат ключа.
- `409` — ключ сгенерирован для другого пользователя.
- `500` — внутренняя ошибка сервера.

#### **Получение кодовой строки**

Хендлер: `GET /api/user/secret/<key>`.

Хендлер доступен только авторизованному пользователю.

Формат запроса:

```
GET /api/user/secret/<key> HTTP/1.1
Content-Length: 0
```

Возможные коды ответа:

- `200` — успешная обработка запроса.

  Формат ответа:

    ```
    200 OK HTTP/1.1
    Content-Type: application/json
    Authorization: ...
    ...
    
    {
        "secret": "<secret>"
    }
    ```

- `401` — пользователь не авторизован.
- `409` - ключ сгенерирован для другого пользователя.
- `422` - неправильный формат ключа.
- `500` — внутренняя ошибка сервера.

### Конфигурирование сервиса "Мастер ключей"

Сервис поддерживает конфигурирование следующими методами:

- адрес и порт запуска сервиса: переменная окружения ОС `RUN_ADDRESS` или флаг `--a` значение по умолчанию "127.0.0.1:8081"
- адрес подключения к базе данных: переменная окружения ОС `DATABASE_URI` или флаг `--d` **обязательный** флаг - формат "postgres://user:<>@localhost:5432/<db>"

Флаги имеют приоритет перед переменными окружения.

### Журналирование
- для журналирования используется [zap](https://github.com/uber-go/zap)
- файл лога создается в текущей папке

### Запуск
- keymaster скомпилированный на Ubuntu 22.04 jammy расположен в папке cmd/.
- строка запуска
```
 ./keymaster --d "postgres://<username>:<password>@localhost:5432/<db>"
```