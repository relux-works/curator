# Словарь и карта книги

Термины ниже различают намерение, bytes, authority и процесс. Если два
объекта имеют одинаковый hash, это ещё не значит, что они обладают
одинаковыми правами или одинаковой жизнью.

## Декларации и содержимое

| Термин | Значение | Где подробнее |
| --- | --- | --- |
| Context package | Версионированная единица instructions и dependencies с manifest | «Слои контекста» |
| Module | Выбранный текст пакета с class и environment applicability | «Слои контекста» |
| Profile | Root package и overlays, из которых разрешается окружение | «Профили и воспроизводимые источники» |
| Skillfile | Декларация нужных skills/sources; версия не равна версии skill manifest | «Профили и воспроизводимые источники», «MCP» |
| Lock | Зафиксированная resolution с точными identities и digest | «Профили и воспроизводимые источники» |
| Pin | Выбор конкретного содержимого вместо moving input | «Профили и воспроизводимые источники» |
| Root context | Основные инструкции profile, сохраняемые в additive project composition | «Слои контекста» |
| System context | Отдельный class для явно допускаемого system-prompt channel | «Слои контекста» |
| Repository knowledge | Одобренные справочные материалы; не native mutable memory | «Безопасность, permissions и trust» |
| Server digest | Hash normalized SERVER; не raw manifest hash | «MCP» |
| Declaration digest | Hash expanded declaration с writer-local install | «MCP» |

## Владение, доверие и policy

| Термин | Значение | Где подробнее |
| --- | --- | --- |
| Managed home | Home profile × environment с manager-owned surfaces и native state | «Managed home» |
| Seed | Native настройка, перенесённая при provisioning по adapter rule | «Managed home», «Environments и harnesses» |
| Projection | Доставленное представление выбранного context на native surface | «Managed home» |
| Drift | Несовпадение управляемого содержимого с recorded state | «Managed home» |
| Candidate | Найденный проектный item, ещё не authority | «Project context и Werkstatt» |
| Approval | Protected разрешение для конкретного item/digest и policy ceiling | «Project context и Werkstatt» |
| Admission | Повторная проверка authority и содержимого перед delivery/start | «Project context и Werkstatt» |
| Grant | Manager-issued право principal выполнить ограниченное изменение | «Безопасность, permissions и trust» |
| Ceiling | Верхняя граница полномочий; approval не отменяет immutable denies | «Безопасность, permissions и trust» |
| Source exclusion | Подавление native загрузки неутверждённых control inputs | «Environments и harnesses» |
| Native workspace trust | Решение самого harness о project surfaces; не per-item manager approval | «Environments и harnesses» |
| Audit | Проверка известных свойств содержимого; не доказательство отсутствия всех угроз | «Безопасность, permissions и trust» |
| Qualification | Измерение заданного контракта на runtime/version/platform и lifecycle | «Environments и harnesses» |

## Процессы и сессии

| Термин | Значение | Где подробнее |
| --- | --- | --- |
| Harness | Программа сессии, которая обращается к модели и исполняет tool calls | «Для кого эта книга» |
| Environment | Зарегистрированный вариант доставки окружения harness | «Environments и harnesses» |
| Adapter | Manager code с native layout и channel facts | «Environments и harnesses» |
| Fragment | Вклад Curator в будущий launch, а не полный process plan | «Environments и harnesses» |
| Construction site | Общее место сборки runtime argv/env для потребителей fragment | «Запуски и жизнь сессии» |
| Werkstatt | Stable project/checkout × profile × environment home с approved generation | «Project context и Werkstatt» |
| Worktree | Git checkout; не manager authority и не домашний каталог harness | «Project context и Werkstatt» |
| Session manager | Интеграция host lifecycle: launch, attach, resume, status | «Запуски и жизнь сессии» |
| Delegated | Caller-side manager composes процесс с live managed home | «Запуски и жизнь сессии» |
| Exact plan | Полный sealed process, связанный с preparation и snapshot | «Запуски и жизнь сессии» |
| Snapshot | Curator-materialized private home и проверяемые bindings | «Запуски и жизнь сессии» |
| Session lease | Удержание original home на время session generation | «Запуски и жизнь сессии» |
| Capture lease | Удержание exact composition на время создания snapshot | «Запуски и жизнь сессии» |
| Detach | Отсоединение терминала с сохранением session lifetime | «Запуски и жизнь сессии» |
| Typed resume | Продолжение прежней identity по closed restart recipe | «Запуски и жизнь сессии» |
| Doorbell | Prepared control-plane notification; не project executable hook | «Запуски и жизнь сессии» |

Определения называют роли, а не готовность всей системы. Например, наличие
слова lease в словаре не утверждает, что любой текущий `profile sync`
удерживает новый project session lease. Для статуса конкретной возможности
нужно читать её главу.

## Карта готовности

| Область | Что доступно или принято | Граница |
| --- | --- | --- |
| Profile → lock → managed home | Resolution, materialization, status и explicit repair ⟦работает сегодня⟧ | Per-adapter residuals и native writable state сохраняются |
| Profile MCP packages | Проверка деклараций, native-shaped files и fragment channels ⟦работает сегодня⟧ | Не все environments имеют channel; не общий strict exclusion |
| Project control context | Protected per-item approval, additive composition и stable Werkstatt ⟦решено⟧ | Реализация и qualification не заменяются проектным Skillfile |
| Project permissions и memory | Manager default deny, grants, native memory off ⟦решено⟧ | Legacy native/yolo не доказывает этот контракт; private memory отложена |
| Project MCP schema 3 | Сначала packages, затем approved inline ⟦решено⟧ | Current readers не принимают новый формат; runtime pin grammar открыта |
| Hosted successor | Exact snapshot-bound attached/detached sessions и doorbells ⟦решено⟧ | Нужны опубликованные carriers, real receiver и совместная qualification |
| Selective exclusion probes | Измеренные controls для конкретных версий ⟦работает сегодня⟧ | Ни один результат не квалифицирует полную strict tuple |

## Указатель рисунков

Каждая схема отвечает на один вопрос. Изображения иллюстрируют отмеченные
в главах контракты; стрелка не является свидетельством реализации.

| Рисунок | Глава | Что читать по схеме |
| --- | --- | --- |
| Порядок слоёв контекста | «Слои контекста» | Profile first, затем approved project general и environment-specific layers |
| От профиля к home | «Профили и воспроизводимые источники» | Декларация, resolution, lock и materialization |
| Project admission | «Project context и Werkstatt» | Candidate не становится authority без protected approval |
| Werkstatt | «Project context и Werkstatt» | Stable home, mutable native state, composition и lease |
| Маршрут запуска | «Project context и Werkstatt» | Delegated live home и цель exact snapshot |
| Runtime exclusion | «Environments и harnesses» | Измеренные positive/negative surfaces и незакрытая qualification |
| Сравнение запусков | «Запуски и жизнь сессии» | Кто composes и какой home потребляет процесс |
| MCP delivery | «MCP» | Package-first и inline admission перед writer |
| Trust boundaries | «Безопасность, permissions и trust» | Кто подаёт данные, кто одобряет и кто исполняет |

## Где искать точный контракт

Public implementation и CLI находятся в
[репозитории Curator](https://github.com/relux-works/curator).
[Спецификация Curator](https://github.com/relux-works/curator-spec) описывает
закрытые форматы и evolution contracts; названия environments и manager
protocols позволяют найти соответствующие разделы. Qualification-документы в `docs/qualification/` репозитория Curator
содержат measured cases и ограничения runtime probes. Команды сборки
книги находятся в `docs/book/README.md`. Эти ссылки не заменяют status
marks в тексте.

Для вопроса «почему загрузился этот файл?» начинают с home, fragment
и native discovery. Для вопроса «кто разрешил действие?» — с policy,
approval и grant. Для вопроса «почему продолжилась именно эта сессия?» —
с identity, provenance и leases. Это разные цепочки доказательств.
