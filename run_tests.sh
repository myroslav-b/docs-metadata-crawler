#!/bin/bash

# Запустити тести і згенерувати покриття для кореневого пакета
go test -v -coverprofile=coverage.out ./...

# Якщо тести пройшли успішно, показати покриття у вигляді HTML
if [ $? -eq 0 ]; then
    go tool cover -html=coverage.out -o coverage.html
    echo "Тести успішно виконані. Звіт про покриття збережено у coverage.html"
    
    # Показати загальний відсоток покриття
    go tool cover -func=coverage.out
else
    echo "Тести завершились з помилками"
fi