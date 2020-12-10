# tail-based-sampling

# Usage

## Start Agent
```
chmod +x ./release/tail-based-sampling
SERVER_PORT=8000 ./release/tail-based-sampling
SERVER_PORT=8001 ./release/tail-based-sampling
```

## Start Collector
```
chmod +x ./release/tail-based-sampling
SERVER_PORT=8002 ./release/tail-based-sampling
```

# Docker build
```
go mod tidy \
    && go mod vendor \
    && docker build -t tail-based-sampling:latest .
```
