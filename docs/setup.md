# Как запустить

## Системные требования

- go 1.25+
- justfile + pwsh (optional)
- docker

## Перед запуском

### Скачать и деархивировать датасеты

Из корня репозитория, в PowerShell на Windows:

```powershell
New-Item -ItemType Directory -Path dataset -Force | Out-Null
curl.exe --fail --location --output dataset/nmap-service-probes https://raw.githubusercontent.com/nmap/nmap/master/nmap-service-probes
curl.exe --fail --location --output dataset/legitimate.zip https://downloads.openappsec.io/waf-comparison-project/legitimate.zip
curl.exe --fail --location --output dataset/malicious.zip https://downloads.openappsec.io/waf-comparison-project/malicious.zip
Expand-Archive -LiteralPath dataset/legitimate.zip -DestinationPath dataset -Force
Expand-Archive -LiteralPath dataset/malicious.zip -DestinationPath dataset -Force
```

### Установить зависимости

```sh
go mod download
go mod verify
just test
```

## Запуск 

### Наивный (baseline) бенчмарк

```sh
just baseline 1 100000
just baseline 4 100000
just baseline 12 100000
just baseline 24 100000
```

Более подробно о бенче в [соответствующем документе](./benchmark.md).

### Hyperscan

```sh
just docker-build
just docker-test
just hyperscan 1 100000 common
just baseline-docker 1 100000 common
```

2 режима работы - common и all, common - берём те регулярки, которые компилируется в Go + работают в Hyperscan, all - берём те, что работают в Hyperscan.

Полный проход со всеми совместимыми Hyperscan паттернами:

```sh
just hyperscan 12 0 all
```

Более подробно о бенче в [соответствующем документе](./benchmark.md).