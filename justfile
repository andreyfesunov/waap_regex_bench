baseline workers="1" limit="100000":
    go run ./cmd/baseline -workers {{workers}} -limit {{limit}}

test:
    go test ./...

race:
    go test -race ./internal/...
