# Справочник CLI и рабочие сценарии

CLI разделяет декларации, resolution, materialization и launch. Команда,
которая обновляет sources, не обязана активировать новый home; read-only
status не должен превращаться в repair. Для автоматизации следует выбирать
конкретную операцию и проверять её exit status.

## Дерево встроенных команд

Дерево ниже получено из `commandRegistry` в `cmd/curator/command_help.go`.
Узлы `<profile>` — operands, а не буквальные имена. Показаны command paths,
без полного списка flags. `help` — отдельная встроенная форма поиска
справки: `curator help <command path>`; она не требует загруженной machine
configuration. ⟦работает сегодня⟧

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

`skill check` обслуживается встроенным checker; другие skill operands
могут передаваться provider. `run` всегда обращается к установленному
provider. Для неизвестного command name допустим umbrella lookup
`curator-<name>` при соблюдении trust rules. Поэтому это дерево исчерпывает
статический реестр, но не перечисляет все устанавливаемые providers.
⟦работает сегодня⟧

Точные flags проверяются, например, так:

```sh
curator help profile install
curator env resolve --help
curator help config build-https
```

Help обнаруживает самый глубокий command path и останавливается на `--`.
Текст в operands после границы не становится запросом help Curator.
Документация старых вариантов flag parsing может отставать от этого
реестра; installed binary нужно проверять отдельно. ⟦работает сегодня⟧

## Какие команды что меняют

Следующая таблица описывает реализацию Curator. Обещание dry run относится
к конкретной команде: оно не означает отсутствие любых network reads
во всех операциях resolution. ⟦работает сегодня⟧

| Группа | Чтение и проверка | Изменение |
| --- | --- | --- |
| Project skills | `status`, `list`, `skill check` | `init`, `add`, `remove`, `install`, `upgrade` |
| Sources | `project resolve` проверяет и фиксирует closure | `project refresh` явно обновляет source lock; `update` fetch-ит sources |
| Global/hybrid | `list`, `status` соответствующей группы | Декларации и installation в своей scope; global также `adopt` |
| Profiles | `profile list`, `profile compose <profile> list` | `install`, `use`, `sync`, `update`, `import`, `remove`, composition add/remove |
| Environments | Обычный `env resolve`, `env status`, `env config show` | Resolve с repair, config set/unset, explicit migration и unmanage |
| Audit/trust | Обычный `audit`, `hook approvals` | Explicit audit allow/publish и hook approve/revoke |
| Machine | `config show`, `version`, help | `bootstrap`, credential configuration, `shell-init --install`, `gc` |
| Sessions | `help run` | Provider-defined launch; capability и prerequisites проверяет provider |

Слова `project add/resolve/refresh` относятся к существующему dependency
workflow. Они не являются полной системой project control admission.
Команды register/inspect/approve для нового project context, показанные
в принятом дизайне, пока не входят в это дерево. Аналогично `env materialize`
для snapshots нельзя выдавать за доступную встроенную команду.
⟦решено⟧

## Сценарий: установить профиль и проверить home

Пусть в публичном source package есть profile `review` с tag `v1.2.3`.
Адрес ниже — placeholder, его заменяют реальным доверенным source.
Первый шаг нужен один раз для machine configuration; он не импортирует
credentials и не логинит harness. ⟦работает сегодня⟧

```sh
curator bootstrap --skills-root ./skill-sources
curator profile install https://code.example.invalid/context/review.git \
  --tag v1.2.3 --as review
curator profile list
curator env resolve codex --profile review --repair --preserve-lock
curator env status --check --json
curator env resolve codex --profile review --format json
```

Install фиксирует выбранный профиль. Repair создаёт или восстанавливает
managed home в пределах lock, затем status и resolve проверяют результат.
Здесь нет `profile use`: проверка named home сама по себе не должна
переопределять machine-current profile. Для осознанного переключения
существует `curator profile use review`. ⟦работает сегодня⟧

## Сценарий: изменить project dependency

Обычный Skillfile управляет skills checkout. Сначала правят декларацию,
затем проверяют план и применяют его. Здесь подразумевается уже созданный
проект с доступными источниками, а не обещание, что install без сети
получит отсутствующее содержимое. ⟦работает сегодня⟧

```sh
curator project resolve .
curator install --dry-run
curator install
curator status --check --json
```

Для source workflow refresh явно заменяет frozen inputs; отдельный install
применяет их. Status не обновляет moving ref по собственной инициативе.
Если нужен fetch плюс install выбранной closure, для этого существует
`upgrade`; `update` только обновляет source repositories. ⟦работает сегодня⟧

```sh
curator project refresh .
curator install --dry-run
curator install
```

Это ещё не одобрение project instructions, MCP или permissions. Наличие
Skillfile рядом с `AGENTS.md` не делает последний admitted control item.
Граница разбирается в главе «Project context и Werkstatt».

## Сценарий: drift и explicit recovery

Когда generated file отредактирован вручную, сначала читают диагностику.
Если нужна исходная версия, repair должен опираться на существующий lock,
а не превращать неисправность в незапрошенное обновление dependency.
⟦работает сегодня⟧

```sh
curator env status --check --json
curator env resolve claude --profile review --repair \
  --preserve-lock --dry-run
curator env resolve claude --profile review --repair --preserve-lock
```

Dry run показывает допустимый repair. Он не утверждает, что любой foreign
file можно присвоить: consent и ownership остаются отдельными. `--takeover`
разрешён лишь в предусмотренных mutation paths и не является универсальным
способом убрать любой refusal. Для credential migration существуют
отдельные `env migrate --inspect`, `--plan` и `--apply`; repair не заменяет
осознанную migration. ⟦работает сегодня⟧

## Сценарий: project env file

Shell hook trust относится к project environment file, а не ко всем
инструкциям проекта. После просмотра конкретного файла оператор может
создать content-bound approval; при изменении bytes требуется новый.
⟦работает сегодня⟧

```sh
curator hook approve .agents/env.sh
curator hook approvals
curator hook revoke .agents/env.sh
```

Эти команды не одобряют MCP endpoint, permission settings или executable
context package. Отзыв shell-hook approval закрывает свой источник, а не
удаляет произвольное native state уже работающего harness.

## Диагностика как часть контракта

| Симптом | Смысл | Дальнейшее действие |
| --- | --- | --- |
| Machine config отсутствует | Group mutation не имеет настроенного manager state | Выполнить осознанный bootstrap, затем повторить |
| `environment_store_untrusted` | Store boundary нельзя считать manager-owned | Исследовать ownership/type/path; не заменять ошибку пустым store |
| `mcp_seed_unstripped` | Codex home сохранил старую base с native MCP | Проверить план миграции; не заявлять чистую базу |
| `environment_passthrough_unreadable` | Currency link нельзя установить | Разобраться с чтением и target; не описывать это как detached |
| `project_view_busy` | Новый проектный контракт удерживает home lease | Ждать или завершать owning session по её lifecycle; пока это принятый следующий слой |

Первые четыре строки относятся к существующим путям Curator; последняя —
к принятому project-aware контракту. По одному сходству кодов нельзя
считать legacy command потребителем нового carrier. Конкретный launcher
и host обязаны объявлять поддерживаемые versions. ⟦решено⟧
