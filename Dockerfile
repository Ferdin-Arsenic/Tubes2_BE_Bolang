FROM golang:1.24.2

WORKDIR /app

COPY src/ ./

RUN go mod tidy

CMD ["go", "run", "treebuilder.go", "bfs.go", "dfs.go", "bidirectional.go", "main.go"]
