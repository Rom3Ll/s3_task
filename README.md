# s3_task

Локальный S3-совместимый стек на [MinIO](https://min.io/) через Docker Compose.

## Требования

- Docker с поддержкой Compose v2 (`docker compose`).

## Быстрый старт

1. Скопируйте пример переменных (опционально; в `docker-compose.yml` уже есть значения по умолчанию):

   ```bash
   cp .env.example .env
   ```

2. Запустите MinIO и одноразовый init, который создаёт бакет:

   ```bash
   docker compose up -d
   ```

3. Проверка:
   - **S3 API:** `http://localhost:9000` (порт задаётся `MINIO_API_PORT`, по умолчанию `9000`).
   - **Веб-консоль MinIO:** `http://localhost:9001` (порт `MINIO_CONSOLE_PORT`, по умолчанию `9001`). Войдите с теми же учётными данными, что в `.env` / `.env.example` (`MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD`).

После первого запуска контейнер `minio-init` создаёт бакет с именем из `S3_BUCKET` (по умолчанию `images`). Повторные запуски безопасны: используется `mc mb … --ignore-existing`.

Остановка и удаление контейнеров (данные MinIO остаются в volume `minio_data`):

```bash
docker compose down
```

Полная очистка включая данные бакета:

```bash
docker compose down -v
```

## Переменные окружения для приложения (Go)

Приложение на хосте обычно ходит в MinIO так:

| Переменная | Пример | Примечание |
|------------|--------|------------|
| `S3_ENDPOINT` | `http://localhost:9000` | С хоста — `localhost`; из другого контейнера в той же сети Compose — `http://minio:9000`. |
| `S3_REGION` | `us-east-1` | Для MinIO часто подходит произвольное значение; главное, чтобы совпадало с тем, что ожидает SDK. |
| `S3_ACCESS_KEY` | совпадает с `MINIO_ROOT_USER` | Например `minioadmin`. |
| `S3_SECRET_KEY` | совпадает с `MINIO_ROOT_PASSWORD` | Например из `.env`. |
| `S3_BUCKET` | `images` | Должен совпадать с бакетом, который создаёт init (переменная `S3_BUCKET` в Compose / `.env`). |
| `S3_USE_PATH_STYLE` | `true` | Типично для MinIO и кастомных endpoint’ов. |
| `LOCAL_SOURCE_DIR` | `assets/source` | Каталог с исходными изображениями для `POST /api/upload-local` (пути `.jpg`/`.jpeg`/`.png`/`.gif`). |
| `LOCAL_UPLOAD_PREFIX` | `uploads` | Префикс ключей в бакете: объекты вида `<префикс>/<имя>.jpg` после перекодирования в JPEG. |

## Загрузка локальных изображений

1. Положите файлы в `assets/source/` (или в каталог из `LOCAL_SOURCE_DIR`).
2. Запустите приложение и откройте UI.
3. Нажмите **«Сжать и загрузить»** — вызовется `POST /api/upload-local`: каждый поддерживаемый файл декодируется, перекодируется в **JPEG с качеством 75** и загружается в бакет под префиксом `uploads/` (или `LOCAL_UPLOAD_PREFIX`).
4. Список объектов обновится с префиксом `uploads/`, чтобы сразу увидеть новые ключи.

Пример для локальной разработки на хосте:

```bash
export S3_ENDPOINT=http://localhost:9000
export S3_REGION=us-east-1
export S3_ACCESS_KEY=minioadmin
export S3_SECRET_KEY=minioadmin
export S3_BUCKET=images
export S3_USE_PATH_STYLE=true
```

Если позже приложение тоже запускается в Docker Compose в одной сети с MinIO, задайте `S3_ENDPOINT=http://minio:9000` и те же ключи/бакет.
