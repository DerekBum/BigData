## Сервер - ячейка памяти

Для запуска сервера нужно из корня проекта вызвать:
```
go run ./server.go <args>
```

Аргументы: 
1) ```port``` отвечает за порт, на котором он будет работать.
Формат: ```:dddd```.
По умолчанию равен ```:8081```.

Сервер будет запущен на localhost-е.

Запоминание последнего ```/replace``` происходит в файл ```db.log```.

Также были реализованы тесты на базовую функциональность.
Для их запуска нужно из корня проекта вызвать:
```
go test -v ./server_test.go ./server.go
```