# Managed home

Managed home — рабочий каталог состояния конкретного harness, подготовленный
Curator из профиля. В нём сочетаются сгенерированные входы и файлы, которые
дальше изменяет сам harness. Главное свойство такого home — известное
владение каждой управляемой поверхностью, а не обещание неизменности
всего каталога.

## Где проходит граница

Менеджер хранит profiles, immutable store и environments под своим
state root. Действующий managed home выбирается по паре profile ×
environment. Следующая схема показывает относительную структуру;
`<state>` обозначает root менеджера, а не каталог конкретного пользователя.
Это сокращённый пример для Codex, не перечень всех служебных файлов.
⟦работает сегодня⟧

```text
<state>/
  profiles/
    review/
      source.json
      lock.json
  contexts/
    context/<package>/<pin>/
    skill/<package>/<pin>/
    skill-context/core-3.1-v1/v<hash-version>/<digest>/
  environments/
    review/
      codex_cli/
        AGENTS.md
        skills/
        config.toml
        .agent-environment.json
```

`source.json` хранит источник установленного профиля, `lock.json` —
его разрешённый выбор. Store entries разделяются profiles, когда member
и pin совпадают. `environments/review/codex_cli/` — место, где Codex
накапливает собственное состояние при запуске с соответствующим home.
Состав файлов зависит от declarations и предыдущих операций: например,
профиль без root surface не обязан иметь `AGENTS.md`. ⟦работает сегодня⟧

Нельзя заменить в этой схеме имя проекта и получить существующий
project-aware home. В действующей функции выбора пути checkout не входит
в ключ. Для persistent checkout × profile × environment home принято
отдельное понятие Werkstatt. ⟦решено⟧

## Adapter переводит понятия в пути

Adapter связывает общий протокол с файловой моделью harness. Он знает,
какая переменная указывает home, куда записывать root instructions,
какие формы допустимы и как передать дополнительные launch channels.
Curator содержит следующие mappings; наличие mapping не подтверждает
все сценарии запуска внешнего consumer. ⟦работает сегодня⟧

| Environment | Выбор home | Root instructions относительно home |
| --- | --- | --- |
| `claude_code` | `CLAUDE_CONFIG_DIR` | `CLAUDE.md` |
| `codex_cli` | `CODEX_HOME` | `AGENTS.md` |
| `muse` | Четыре managed XDG roots | `config/muse/AGENTS.md` |
| `pi` | `PI_CODING_AGENT_DIR` | `AGENTS.md` |
| `opencode` | Managed `XDG_CONFIG_HOME`, home в дочернем `opencode/` | `AGENTS.md` внутри этого home |

Например, managed root Claude копируется как regular file, тогда как
другие поверхности могут ссылаться на immutable store. Muse использует
раздельные config, data, state и cache roots; знание только одного
`XDG_CONFIG_HOME` неполно. В OpenCode переменная указывает родителя,
а сам harness читает дочерний каталог. ⟦работает сегодня⟧

## Четыре класса файлов

Чтобы понять `repair`, сначала надо разделить содержимое home.

| Класс | Кто определяет bytes | Что делает менеджер |
| --- | --- | --- |
| Generated surfaces | Profile lock и adapter | Публикует, фиксирует identity, проверяет drift |
| Provisioning seeds | Начальное состояние из разрешённого native source либо manager-written seed | Создаёт первоначально, затем сохраняет tool-owned state |
| Credential entries | Отдельная credential strategy | Проверяет стратегию и допустимые links; не переносит token bytes как profile content |
| Native mutable state | Harness | Сессии, история и собственные изменения вне generated ownership |

Generated surface может быть root instruction file, skill projection
или MCP configuration. Marker `.agent-environment.json` связывает её
с профилем, lock hash, формой, путями и content hashes. Marker описывает
владение и ожидаемое состояние; он не является секретом, login или
универсальным разрешением исполнить всё в каталоге. ⟦работает сегодня⟧

## Seeds: один раз не значит при каждом repair

Provisioning seed помогает harness начать работу с привычной конфигурацией.
В Codex `config.toml` первоначально берётся из native home. Действующее
преобразование удаляет top-level `mcp_servers` вместе с subtables, чтобы
не унаследовать эти servers как неявную managed configuration. Остальные
TOML values сохраняются по правилам seed reader. ⟦работает сегодня⟧

После provisioning этот `config.toml` принадлежит harness. Если пользователь
изменил там настройку модели, обычный repair не должен заново принести
старое значение из native source. Изменение исходного seed также не
является автоматическим обновлением managed copy. Нечитаемый native
copy seed блокирует первоначальное чтение, но не операцию repair,
которая этот seed вообще не перечитывает. ⟦работает сегодня⟧

У Claude `.claude.json` создаётся менеджером с нужной launch-directory
trust записью. При последующем repair Curator меняет только необходимые
flags этой записи и сохраняет посторонние значения. Уже удовлетворённая
запись не переписывается. Это semantic JSON merge: если изменение нужно,
порядок ключей и whitespace могут измениться, хотя неизвестные значения
и числовые lexemes сохраняются. ⟦работает сегодня⟧

Например, harness добавил в `.claude.json` собственный счётчик, а затем
потребовалось разрешённое включение referenced context. Repair может
добавить требуемый flag, но не вправе сбросить счётчик, заменив весь
документ исходным seed. Повреждённый или symlinked файл не считается
пустым новым home: действуют отдельные отказы чтения. ⟦работает сегодня⟧

## Skill projection отделяет контекст от исходников

Store может хранить выбранный skill package целиком. В discovery directory
harness должен попадать только предназначенный агенту контекст. Curator
строит projection по правилам context selection и публикует её отдельно;
managed и native skill links ссылаются на эту projection, а не на сырой
repository tree. ⟦работает сегодня⟧

В частности, repository-only материалы и исключённые runtime/build roots
не должны проникать в installed context только потому, что оказались рядом
с `SKILL.md`. Command publishers читают свои snapshots и используют
отдельные runtime stores и build caches. Контекст skill и исполняемый
артефакт связаны декларацией, но доставляются разными путями.
⟦работает сегодня⟧

Projection имеет content-addressed identity. Отсутствующий store,
изменённые bytes, лишние или пропавшие файлы делают её stale. Explicit
repair перестраивает projection из pinned store entry и заменяет дерево
через recoverable transaction после проверки identity. Обновить baseline
на произвольное текущее содержимое — не repair. ⟦работает сегодня⟧

## Проверка, repair и recovery

Обычный `env resolve` проверяет ожидаемое состояние. При drift он сообщает
`environment_home_stale` с причинами и не выдаёт успешный fragment.
`--repair` разрешает provisioning или восстановление под manager-home
mutation lock; dry-run сообщает потребность, не публикуя изменения.
⟦работает сегодня⟧

```sh
curator env resolve codex_cli --profile review --format json
curator env resolve codex_cli --profile review --repair --preserve-lock
curator env status --check --json
```

`--preserve-lock` требует `--repair`. Он нужен, когда effective pin должен
остаться прежним: если repair потребует identity migration lock,
операция откажется до locking и writes. Обычный repair без этого флага
может выполнить предусмотренную миграцию. Это узкая гарантия lock,
а не требование оставить все stale surfaces нетронутыми.
⟦работает сегодня⟧

Recovery отвечает на другой вопрос: что делать с незавершённой публикацией
после прерывания. Менеджер использует journaled transactions со staging
и atomic rename; последующая операция восстанавливает согласованное
состояние. Mutation lock сериализует работу менеджера, но не делает
harness и другие same-user процессы участниками этой транзакции.
⟦работает сегодня⟧

Foreign файл нельзя молча объявить своей generated surface. Операции,
которым нужен takeover, требуют соответствующего согласия и сохраняют
backup по контракту поверхности. Ошибка чтения ownership evidence не
равна отсутствию файла. Именно эта разница не позволяет «починить»
непонятное состояние поверх чужих данных. ⟦работает сегодня⟧

## Login не является частью profile package

Credential strategy задаётся отдельно от instructions. Возможны reuse
native login через допустимый link либо собственный login изолированного
home; применимость зависит от adapter, платформы и backend. Curator не
копирует token bytes в context store, profile lock или generated
instructions. ⟦работает сегодня⟧

Например, для Claude на macOS shared strategy в действующем adapter
не поддерживается; отдельный home требует собственного login. Для Codex
допустимость изоляции зависит от native credential backend. Переключение
машинного selector не следует считать готовой миграцией login: часть
runtime mappings ещё отключена и должна отказать, а не подставить
похожий вариант. ⟦работает сегодня⟧

## Live home и snapshot сессии

Managed home изменяется со временем, в том числе самим harness. Принятый
snapshot design требует выделять сессии проверенный самостоятельный набор
входов; это не архив всего home вместе с credentials. В Curator уже есть
внутренние части capture и snapshot verification/publication, но их наличие
не означает готовую end-to-end CLI и launch integration. ⟦решено⟧

Эта граница важна для следующей главы. Persistent home экономит стоимость
повторной подготовки и сохраняет native state, а неизменная composition
и lease должны удержать конкретные управляющие входы запуска. Один
механизм нельзя заменить обещанием другого.
