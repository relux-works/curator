# План книги «Curator: окружение агента»

Книга идёт от различения данных и управляющего контекста к установке,
проверке и запуску. Каждая глава читается самостоятельно после главы
«Слои контекста». Термины вводятся по роли, без истории разработки.

| Глава | Вопрос читателя | Содержание | Наличие текста |
| --- | --- | --- | --- |
| Для кого эта книга | Что здесь объясняют? | Термины, границы, статусы и маршрут чтения | Полная глава |
| Зачем управлять окружением | Почему одного файла инструкций мало? | Скрытые входы, воспроизводимость, данные и полномочия | Полная глава |
| Слои контекста | Что именно получает агент и в каком порядке? | Context package, profile root, additive project composition и native discovery | Полная глава |
| Профили и воспроизводимые источники | Как намерение становится проверяемым набором? | Manifest, Skillfile, sources, lock, overlays, audit | Полная глава |
| Managed home | Кто владеет файлами и что восстанавливается? | Layout, seeds, projections, drift, repair, recovery, credentials | Полная глава |
| Project context и Werkstatt | Как принять правила проекта? | Registration, candidates, per-item approval, composition, scope, leases, worktree | Полная глава |
| Environments и harnesses | Как различается native discovery? | Registry, adapters, resolve fragment, seeds, runtime exclusion matrix и measured gaps | Полная глава |
| Запуски и жизнь сессии | Кто собирает процесс и кто им владеет? | Native, delegated, exact plan, snapshots, detach, resume, doorbells, leases | Полная глава |
| MCP | Как сервер становится разрешённым инструментом? | Skill requirements, packages, Stage A/B, digests, native writers, runtime pins | Полная глава |
| Безопасность, permissions и trust | Где заканчивается управление контекстом? | Default deny, grants, project permissions, memory-off, audit, network, signatures и residuals | Полная глава |
| Справочник CLI и рабочие сценарии | Какая команда меняет какое состояние? | Generated command tree, profile/project workflows, update, drift/repair, hook revoke | Полная глава |
| Словарь и карта книги | Где найти определение и статус? | Термины, readiness map и указатель девяти рисунков | Полная глава |

Практические сценарии включены в справочник CLI; материал об устройстве
окружения и процессах разделён на две главы. Private-memory migration,
retention и несуществующие admission CLI не выдаются за рабочие сценарии.

У девяти рисунков отдельные исходники и SHA-256 в `src/MANIFEST.md`:
порядок слоёв, profile → lock → home, project admission, Werkstatt,
маршрут запуска, runtime exclusion, сравнение launch routes, MCP delivery
и trust boundaries. Измерения runtime exclusion дополняются таблицей
в главе об environments.
