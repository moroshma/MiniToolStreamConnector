#!/bin/bash

# Скрипт для создания тегов для multi-module репозитория

set -e  # Остановить при ошибке

echo "=== Шаг 1: Удаление старого тега v0.0.1 ==="
git tag -d v0.0.1 2>/dev/null || echo "Локальный тег v0.0.1 не найден"
git push origin :refs/tags/v0.0.1 2>/dev/null || echo "Тег v0.0.1 на GitHub не найден"

echo ""
echo "=== Шаг 2: Создание новых тегов для каждого модуля ==="
git tag model/v0.0.1
echo "✓ Создан тег model/v0.0.1"

git tag minitoolstream_connector/v0.0.1
echo "✓ Создан тег minitoolstream_connector/v0.0.1"

git tag minitoolstream_connector/subscriber/v0.0.1
echo "✓ Создан тег minitoolstream_connector/subscriber/v0.0.1"

echo ""
echo "=== Шаг 3: Отправка всех тегов в GitHub ==="
git push origin model/v0.0.1
git push origin minitoolstream_connector/v0.0.1
git push origin minitoolstream_connector/subscriber/v0.0.1

echo ""
echo "=== Шаг 4: Проверка созданных тегов ==="
git tag -l

echo ""
echo "✓ Все теги успешно созданы и отправлены в GitHub!"
echo ""
echo "Теперь вы можете проверить работу командой:"
echo "  cd minitoolstream_connector"
echo "  go clean -modcache"
echo "  go get -u github.com/moroshma/minitoolstream_connector/model@v0.0.1"
