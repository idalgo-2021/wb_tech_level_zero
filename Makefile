ENV_FILE_PATH := ./.env
COMPOSE_FLAGS = --env-file $(ENV_FILE_PATH)
COMPOSE_DB_FILE = -f docker-compose.yaml
# PGDATA_DIR := ./pgdata
PGDATA_DIR := postgres_data

# =============================================================================
# БАЗА ДАННЫХ
# =============================================================================

## db-up: Запустить ТОЛЬКО базу данных для локальной отладки
db-up:
	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) up -d db

## db-down: Остановить контейнер с базой данных для отладки
db-down:
	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) down db

## db-logs: Показать логи базы данных для отладки
db-logs:
	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) logs -f db

# =============================================================================
# KAFKA
# =============================================================================

## kafka-up: Запустить только Kafka и Kafka UI
kafka-up:
	docker compose $(COMPOSE_DB_FILE) up -d kafka kafka-ui

## kafka-down: Остановить Kafka и Kafka UI
kafka-down:
	docker compose $(COMPOSE_DB_FILE) down kafka kafka-ui

## kafka-logs: Показать логи Kafka
kafka-logs:
	docker compose $(COMPOSE_DB_FILE) logs -f kafka

# =============================================================================
# REDIS
# =============================================================================

## redis-up: Запустить только Redis
redis-up:
	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) up -d redis

## redis-down: Остановить Redis
redis-down:
	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) down redis

## redis-logs: Показать логи Redis
redis-logs:
	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) logs -f redis


# =============================================================================
# ВСЁ ОКРУЖЕНИЕ
# =============================================================================

## up: Запустить все сервисы (Postgres + Kafka + Kafka UI + Redis)
up:
	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) up -d

## down: Остановить все сервисы
down:
	docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) down

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
		docker compose $(COMPOSE_DB_FILE) $(COMPOSE_FLAGS) down -v && \
		sudo rm -rf $(PGDATA_DIR) && \
		echo "✅ Контейнеры и данные удалены."; \
	else \
		echo "❌ Отменено."; \
	fi
