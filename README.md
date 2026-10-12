# waap_regex_bench

Убедительная просьба отнестись со вниманием к документации, она писалась руками.

## Файловая структура

- [dataset/](dataset/) - папка с исходниками, в том числе [dataset/legitimate/](dataset/legitimate/) - целевой датасет, [dataset/malicious/](dataset/malicious/) - вредоносный, [dataset/nmap-service-probes](dataset/nmap-service-probes) - список regex по nmap.
- [cmd/](./cmd/) - точки входа в приложения для оценки производительности тех или иных методов.
- [internal/](./internal/) - внутренняя логика приложений.
- [bench-results/](./bench-results/) - результаты прогонов бенчмарков. 

## Документация

- [docs/task.md](./docs/task.md) - исходник задания.
- [docs/regex_approaches.md](./docs/regex_approaches.md) - рассуждения и решения на тему работы с регулярными выражениями.
- [docs/stack.md](./docs/stack.md) - рассуждения и решения на тему выбора стека.
- [docs/setup.md](./docs/setup.md) - как запустить проект и бенчмарки.
- [docs/dev.md](./docs/dev.md) - какие подходы использовались при разработке.
- [docs/benchmark.md](./docs/benchmark.md) - результаты работы.
- [docs/hyperscan.md](./docs/hyperscan.md) - как и почему работает система.
- [docs/limitations.md](./docs/limitations.md) - ограничения системы при масштабировании.