# booksearch

Сервер gRPC, выполняющий запросы к базе данных books
MySQL: поиск книг по автору и автора по книге.

Инструкция по запуску:

1. Собрать образ БД c тестовыми значениями с помощью Dockerfile.

```
make ms_build
```

2. Запустить контейнер. Если появляется ошибка "порт занят", на
компьютере, скорее всего, запущена MySQL, которую нужно отключить.

```
make ms_run
```

3. Запустить сервер.
```
make run
```

Примечание. После выполнения шага 2 нужно подождать около 20 секунд,
прежде чем контейнер запустится, иначе может возникнуть ошибка.

4. В другом терминале собрать приложение-клиент.

```
make client
```

5. Примеры запросов клиента:

```
bkclient --author 'Гоголь'
bkclient --book 'Преступление'
```

Удалить клиент-приложение:

```
make bk_clean
```

Остановить контейнер БД:

```
make ms_stop
```

Удалить контейнер и образ:

```
make ms_clean
```
