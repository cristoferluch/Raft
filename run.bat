@echo off
start cmd /k "title localhost:9080 && go run main.go ./configs/config1.yaml"
start cmd /k "title localhost:9081 && go run main.go ./configs/config2.yaml"
start cmd /k "title localhost:9082 && go run main.go ./configs/config3.yaml"
