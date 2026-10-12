# waap_regex_bench

## Файловая структура

- [docs/](docs/) - документация, там лежит исходник задания
- [dataset/](dataset/) - папка с исходниками, в том числе [dataset/legitimate/](dataset/legitimate/) - целевой датасет, [dataset/malicious/](dataset/malicious/) - вредоносный, [dataset/nmap-service-probes](dataset/nmap-service-probes) - список regex по nmap

## Решения

- Проход по 1e6 записей с проверкой каждого из 12000 регулярных выражений. Сразу нет - O(n * k) асимптотика - неэффективно.