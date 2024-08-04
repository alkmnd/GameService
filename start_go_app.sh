#!/bin/bash

# Источник переменных окружения из .env-vars
source /home/2cubes/backend/.env-vars
# Отладочная информация
echo "Starting Go application with the following environment:"
env | grep -E '^(export )?'

# Запуск приложения Go
echo "Executing Go application..."
/usr/bin/go run /home/2cubes/gameservice/cmd/main.go
