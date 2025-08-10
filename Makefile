ENV_FILE_PATH := ./.env
COMPOSE_FLAGS = --env-file $(ENV_FILE_PATH)
COMPOSE_DB_FILE = -f docker-compose.yaml
PGDATA_DIR := ./pgdata

## db-up: Запустить ТОЛЬКО базу данных для локальной отладки
db-up:
##	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) up -d db
	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) up -d

## db-down: Остановить контейнер с базой данных для отладки
db-down:
##	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) down db
	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) down

## db-logs: Показать логи базы данных для отладки
db-logs:
##	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) logs -f db
	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) logs -f

# =============================================================================
# КОМАНДЫ ОЧИСТКИ
# =============================================================================

## prune: Удалить все остановленные контейнеры, неиспользуемые сети и образы
prune:
	docker system prune -af

## clean: Полностью остановить окружение и УДАЛИТЬ ДАННЫЕ БАЗЫ ДАННЫХ
clean:
	@read -p "⚠ Это удалит ВСЕ данные из $(PGDATA_DIR). Продолжить? [y/N] " confirm && \
	if [ "$$confirm" = "y" ]; then \
		docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) down -v db && \
		sudo rm -rf $(PGDATA_DIR) && \
		echo "✅ Контейнер и данные БД удалены."; \
	else \
		echo "❌ Отменено."; \
	fi
