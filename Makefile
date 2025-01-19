MAKEFLAGS += -j2
.PHONY: clean install run clean server tailwind docker-run docker-down

# Run server
run: server tailwind

install:
	@go install github.com/a-h/templ/cmd/templ@latest
	@go install github.com/bokwoon95/wgo@latest
	@npm install -D tailwindcss
	@npm install @tailwindcss/forms && npm install @tailwindcss/typography

# Clean process left by wgo in server command
# This is only needed because the inner process from templ generate does not stop properly
# With `go run .` i works, but with the filewatcher (wgo) it does not kill the inner process
clean:
	@clear; \
	PID=$$(pgrep -f ht-2025-ai); \
	if [ -n "$$PID" ]; then \
		kill -9 $$PID; \
		echo "Killed process $$PID"; \
	else \
		echo "No programm running with name ht-2025-ai"; \
	fi;

# Run server with hot reload and go file watcher
# With --proxy="http://localhost:2323" (which opens the browser tab with the website) it seems
# to be much more prone to the error of not quitting processes.
# server: clean
#	templ generate --watch --cmd 'wgo run . -name ht-2025-ai'
server:
	@wgo -file .go -file .templ -xfile _templ.go clear :: templ generate :: go run . -name ht-2025-ai

# Run tailwind watcher
tailwind:
	@npx tailwindcss -i ./web/static/styles/index.css -o ./web/static/styles/output.css --watch

python:
	@python3 jobs/main.py

# Create DB container
docker-run:
	@if docker compose up 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi