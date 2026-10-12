set windows-shell := ['powershell.exe', '-NoProfile', '-Command']

baseline workers="1" limit="100000":
    go run ./cmd/baseline -workers {{workers}} -limit {{limit}}

test:
    go test ./...

race:
    go test -race ./internal/...

docker-build:
    docker build --platform linux/amd64 -t waap-regex-bench .

hyperscan workers="1" limit="100000" set="all":
    powershell.exe -NoProfile -File docker-bench.ps1 -Engine hyperscan -Workers {{workers}} -Limit {{limit}} -PatternSet {{set}}

baseline-docker workers="1" limit="100000" set="all":
    powershell.exe -NoProfile -File docker-bench.ps1 -Engine baseline -Workers {{workers}} -Limit {{limit}} -PatternSet {{set}}

docker-test:
    powershell.exe -NoProfile -File docker-bench.ps1 -Engine test

docker-environment:
    powershell.exe -NoProfile -File docker-bench.ps1 -Engine environment
