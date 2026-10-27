# Справочник CLI и рабочие сценарии

CLI Curator разделяет объявление, разрешение, материализацию и запуск. Команда, которая обновляет источники, новый дом не активирует, а команда статуса только читает и ничего не восстанавливает. Для автоматизации выбирайте конкретную операцию и проверяйте её код выхода.

## Дерево встроенных команд

Встроенные команды перечисляет реестр `commandRegistry` в `cmd/curator/command_help.go`. Узел `<profile>` обозначает операнд с именем профиля. Дерево реестра без флагов выглядит так:

```text
curator
  bootstrap
  init
  add
  remove
  install
  update
  upgrade
  status
  list
  project
    add
    register
    rebind
    inspect
    package
    preview
    approve
    revoke
    resolve
    refresh
  skill
    check
  global
    init
    add
    remove
    list
    status
    install
    adopt
    update
    upgrade
  profile
    install
    import
    list
    use
    update
    remove
    mcp
      install
      status
      repair
      restore
      replace
      remove
      release
    sync
    compose
      <profile>
        add
        remove
        list
  env
    resolve
    status
    config
      show
      set
      unset
    migrate
    unmanage
  run
  hybrid
    add
    remove
    list
    status
  audit
  gc
  shell-init
  hook
    approve
    approvals
    revoke
  ui
  config
    show
    build-ssh
      add
      list
      remove
    build-https
      add
      login
      list
      remove
  version
```

`skill check` обслуживает встроенная проверка пакета скилла, остальные операнды `skill` уходят провайдеру. `run` всегда передаёт управление установленному провайдеру `curator-run` (глава "Запуски и жизнь сессии"). Перед этим Curator переписывает псевдонимы `claude` и `codex` в канонические идентификаторы `claude_code` и `codex_cli`, а аргументы после `--` передаёт без изменений.

Неизвестное имя команды Curator ищет как исполняемый файл `curator-<name>` в `PATH`, по соглашению внешних подкоманд git. Корнями доверия для такого поиска служат каталог исполняемого файла Curator и список `provider_directories` из конфигурации машины. Совпадение внутри корней доверия Curator запускает молча, а совпадение вне их запускает с предупреждением `subcommand_provider_outside_trust_roots`. Файл в каталоге, который публикует или ведёт сам Curator, отклоняется с `subcommand_provider_untrusted`. Дерево поэтому исчерпывает статический реестр и устанавливаемых провайдеров не перечисляет.

## Справка

`curator help <путь команды>` печатает синопсис и описание команды из того же реестра. Флаги `-h` и `--help` в любом месте до `--` дают ту же справку. Справка не читает конфигурацию машины и ничего не пишет, поэтому работает и на машине без `curator bootstrap`. Точные флаги проверяйте справкой установленной программы:

```sh
curator help profile install
curator env resolve --help
curator help config build-https
```

Справка выбирает самый глубокий подходящий путь команды и останавливается на `--`. Текст операндов после этой границы запросом справки Curator не становится: `curator run codex -- -h` передаёт `-h` провайдеру.

Флаг `--takeover` разрешён только там, где команда умеет взять под управление чужие файлы: в `profile install --use`, `profile use`, `profile sync`, `profile update`, командах `profile mcp` и в `env resolve --repair` без `--dry-run`. В любой другой комбинации Curator отказывает с ошибкой использования до запуска команды.

## Какие команды что меняют

Каждая группа команд разделяет чтение и изменение:

| Группа | Чтение и проверка | Изменение |
| --- | --- | --- |
| Скиллы проекта | `status`, `list`, `skill check` | `init`, `add`, `remove`, `install`, `upgrade` |
| Источники Skillfile | `status` сравнивает установленное с lock | `project resolve` фиксирует источники в `Skillfile.lock.json`, `project refresh` явно обновляет этот lock, `update` получает обновления репозиториев источников |
| Допуск проекта | `project inspect`, `project preview` без `--declaration` | `project register`, `rebind`, `package`, `preview --declaration`, `approve`, `revoke` пишут только защищённое состояние менеджера |
| Глобальные и гибридные скиллы | `list` и `status` своей группы | объявления и установка в своей области, у `global` ещё `adopt` |
| Профили | `profile list`, `profile compose <profile> list`, `profile mcp status` | `install`, `use`, `sync`, `update`, `import`, `remove`, `compose add` и `remove`, `profile mcp install` и `repair` для отсутствующих записей |
| Агентские среды | `env resolve` без `--repair`, `env status`, `env config show`, `env migrate --inspect` и `--plan` | `env resolve --repair`, `env config set` и `unset`, `env migrate --apply`, `env unmanage` |
| Аудит и доверие | `audit`, `hook approvals` | `audit --allow` с `--reason`, `audit --publish`, `hook approve`, `hook revoke` |
| Машина | `config show`, списки `build-ssh` и `build-https`, `version`, `help`, `ui` | `bootstrap`, добавление и удаление учётных данных сборки, `shell-init --install`, `gc` |
| Сессии | `help run` | `run` передаёт запуск провайдеру, который сам проверяет возможности и предпосылки |

Флаг `--dry-run` обещает отсутствие записи только для своей команды. Сетевых чтений при разрешении источников он не исключает. `project resolve` для `Skillfile.json` первой схемы только печатает пути и диск не меняет.

Команды `project add`, `resolve` и `refresh` ведут зависимости `Skillfile.json`. Допуск управляющего контекста проекта ведут другие команды той же группы: `register`, `inspect`, `package`, `preview`, `approve`, `revoke` и `rebind`. Они читают и пишут только защищённое состояние менеджера, композиции не собирают и харнес не запускают.

Действия `profile mcp restore`, `replace`, `remove` и `release` требуют отдельного полномочия владения записями MCP и без него отказывают. `profile mcp install` и `repair` впервые устанавливают отсутствующие записи, а действия владения в них тоже отказывают.

## Сценарий: установить профиль и проверить дом

Пусть публичный пакет источника содержит профиль `review` с тегом `v1.2.3`. Адрес в примере служит заглушкой, его заменяют реальным доверенным источником. Первая команда один раз создаёт конфигурацию машины. Учётных данных она не импортирует и в харнес не входит. В позиции безопасности `hardened` до установки настройте белые списки и ключи подписантов источников (глава "Безопасность, разрешения и доверие"). Последовательность команд такая:

```sh
curator bootstrap --skills-root ./skill-sources
curator profile install https://code.example.invalid/context/review.git \
  --tag v1.2.3 --as review
curator profile list
curator env resolve codex --profile review --repair --preserve-lock
curator env status --check --json
curator env resolve codex --profile review --format json
```

`profile install` проверяет каждый член профиля строгим аудитом и фиксирует выбранный профиль в lock. Если текущего профиля на машине нет, первая успешная установка активирует `review` и без `--use`. Если текущий профиль уже есть, установка без `--use` его сохраняет, а явное переключение делает `curator profile use review`.

`env resolve --repair` создаёт или восстанавливает управляемый дом в пределах lock. С `--preserve-lock` восстановление отказывает до записи, если оно изменило бы хэш lock. Затем `env status` и `env resolve` проверяют результат. `env resolve` с `--profile review` проверяет дом названного профиля и текущий профиль машины не переключает.

## Сценарий: изменить зависимость проекта

`Skillfile.json` проекта управляет скиллами чекаута. Сначала правят декларацию, затем фиксируют источники, проверяют план и применяют его. Сценарий предполагает уже созданный проект и доступные источники: без сети `install` отсутствующего содержимого не получит. Команды такие:

```sh
curator project resolve .
curator install --dry-run
curator install
curator status --check --json
```

`Skillfile.json`, изменённый после `project resolve`, отказывает с `source_lock_stale` до явного обновления lock. Обновление выглядит так:

```sh
curator project refresh .
curator install --dry-run
curator install
```

`project refresh` заново разрешает ссылки, байты и состав и пишет новый lock. Сам по себе он ничего живого не меняет, новый lock применяет следующий `install`. `status` подвижную ссылку не продвигает. Получение и установку выбранного замыкания одной командой делает `upgrade`, а `update` только получает обновления репозиториев источников под `skills_root`.

Установка зависимостей не одобряет инструкций проекта, MCP и разрешений. `Skillfile.json` рядом с `AGENTS.md` не делает этот файл допущенным управляющим элементом (глава "Контекст проекта и рабочее место").

## Сценарий: допустить инструкции проекта

Чтобы `AGENTS.md` проекта стал одобренным элементом, оператор регистрирует чекаут, упаковывает файл, рассматривает пакет и одобряет его. Нужны установленный профиль и чекаут Git хотя бы с одним коммитом, а дом менеджера лежит вне чекаута. Команды печатают JSON, и идентификатор чекаута и дайджесты берут из их вывода:

```sh
curator project register ./demo --name demo --profile dev --environment codex_cli
curator project inspect CHECKOUT_ID
curator project package CHECKOUT_ID instructions --source AGENTS.md
curator project preview CHECKOUT_ID
curator project approve CHECKOUT_ID instructions \
  --digest ITEM_SHA256 --package-digest REVIEW_SHA256
```

`register` выделяет чекауту случайный идентификатор менеджера и связывает его с корнем Git, сетевой идентичностью origin, корневым коммитом, устройством и inode каталога, lock профиля и агентской средой. Репозиторий команда не меняет. Регистрация из CLI получает строгий режим. `inspect` сообщает режим отдельно от допуска элементов (`admission: "off"`) и памяти (`memory: "off"`). Привязка к файловой системе работает на macOS и Linux, на других платформах регистрация отказывает.

`package` явно выбирает один найденный `AGENTS.md` или `CLAUDE.md`, сохраняет его байты LF без изменений и хранит снимок вне репозитория. Обнаружение само такого файла не выбирает. `preview` выводит пакет на рассмотрение и его дайджест, а `approve` записывает одобрение ровно этого элемента в пределах политики менеджера. Изменённый файл требует новой упаковки и нового рассмотрения. Отзывает одобрение `curator project revoke CHECKOUT_ID instructions`, и отзыв работает даже без чекаута на диске.

Перемещённый, заменённый или нечитаемый чекаут требует явного восстановления привязки:

```sh
curator project rebind CHECKOUT_ID ./demo
```

`rebind` выводит старую запись из работы и выделяет новый идентификатор, одобрения при этом не переносятся. Возврат прав доступа или того же inode старую регистрацию не оживляет.

## Сценарий: дрейф и явное восстановление

Когда сгенерированный файл отредактирован вручную, сначала читают диагностику. Если нужна исходная версия, восстановление опирается на существующий lock и не превращает неисправность в незапрошенное обновление зависимостей:

```sh
curator env status --check --json
curator env resolve claude --profile review --repair \
  --preserve-lock --dry-run
curator env resolve claude --profile review --repair --preserve-lock
```

При дрейфе сгенерированного файла пробный запуск с `--dry-run` файлов не пишет и завершается с кодом 1 и `environment_home_stale`. Он подтверждает устаревшее состояние и плана восстановления не печатает. Восстанавливает дом следующий запуск без `--dry-run`.

Согласие и владение остаются отдельными условиями. `--takeover` работает только в перечисленных выше командах и не снимает любой отказ. Перенос учётных данных выполняет отдельная команда `env migrate`, и восстановление её не заменяет. `--inspect` печатает опись только для чтения, `--plan` печатает точные операции с хэшем плана, а `--apply` требует `--expect` с этим хэшем:

```sh
curator env migrate --plan --env codex_cli
curator env migrate --apply --env codex_cli --expect PLAN_HASH
```

`--apply` отказывает, если опись изменилась после `--plan` или остались конфликты. Байтов учётных данных миграция не копирует, не перемещает и не переписывает.

## Сценарий: файл окружения проекта

Доверие к хуку оболочки относится к одному файлу окружения проекта и на инструкции проекта не распространяется. После просмотра конкретного файла оператор записывает одобрение, связанное с его содержимым. Изменённые байты требуют нового одобрения. Команды такие:

```sh
curator hook approve .agents/env.sh
curator hook approvals
curator hook revoke .agents/env.sh
```

Эти команды не одобряют MCP-сервер, настройки разрешений и исполняемый контекстный пакет. Отзыв одобрения хука закрывает только свой источник и нативного состояния уже работающего харнеса не трогает.

## Диагностика

Код диагностики входит в контракт команды. Частые коды означают следующее:

| Симптом | Смысл | Действие |
| --- | --- | --- |
| Конфигурации машины нет | Команде группы нужна конфигурация менеджера, и Curator подсказывает `curator bootstrap --skills-root <dir>` | Осознанно выполнить `bootstrap` и повторить команду |
| `environment_home_stale` | Управляемый дом расходится с lock | Прочитать причины и выполнить `env resolve --repair` |
| `environment_store_untrusted` | Границу хранилища нельзя считать принадлежащей менеджеру | Исследовать владельца, тип и путь; пустым хранилищем ошибку не заменять |
| `mcp_seed_unstripped` | Дом Codex сохранил старую начальную настройку с нативными MCP | Проверить план миграции; чистой базы не заявлять |
| `environment_passthrough_unreadable` | Ссылку на учётные данные управляемого дома нельзя осмотреть или прочитать её цель | Вернуть доступ к записи и её родительским каталогам; отсоединённой эту ссылку не считать |
| `source_lock_stale` | `Skillfile.json` изменился после фиксации источников | Выполнить `project refresh`, затем `install` |
| `project_context_unsupported` | Адаптер не обеспечивает требуемую поверхность, или в режиме `native-project` выбран элемент контекста | Выбрать поколение без этого элемента или квалифицированную связку среды, версии и платформы |

Одинаковый код в разных командах не делает старую команду потребителем нового контракта. Лаунчер и хост сессий объявляют поддерживаемые версии своих форматов сами.
