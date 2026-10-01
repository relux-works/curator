TASK-261001-31dgus carry results

Applied ONLY the .research hunks of both CR patches via git apply --include=.research/*; LOGBOOK.md untouched.

3qugz9 candidate tree 064f083a5b561f1a511d08ac6e3288e5d3e1114d:
- .research/261001_second-operator-requirements-answers.md blob=ece7c542385572039d14bc3a0ec1cbde8e529d63 sha256=90c4d915da11d9eaf5783f35e54cb3e17a4d37541450acd0682401e4eec84970 MATCH
- .research/TASK-261001-3qugz9_capture-provider.py blob=a43d77fe73513cd59f41dd9fe6415c07fda29af7 sha256=e4b5800b9608b45592d6d9118b05f980309227fb80e602fb5ac31e89a940f7ba MATCH
- .research/TASK-261001-3qugz9_evidence.json blob=25bf4627ad357d9a0dbc666288ce51eeb74dad6d sha256=0ab6b48bc76a3840c45adcf888114ed5559d9233e948433df9f51beffbc382a9 MATCH
- .research/TASK-261001-3qugz9_probe.py blob=3e98ef578ea963585f884f84c2904f12405d2d6a sha256=3abe85eac54deff2674a362d331c304a091aa6aff25948cd30d2d269d9fd45f9 MATCH

3s8csu candidate tree 1d3acaf8979f92cb1f80ac9fc85b0cf824854e42:
- .research/261001_mandates-launch-context-advice.md blob=b4e133a458f13b7ac9e104e9944f298447631f3f sha256=c66b4e7524504375d9851a6bdea18bf4e7db9880afb75625e3247557b69658cf MATCH

Each MATCH means git hash-object(file) equals git rev-parse(<tree>:<path>).
LOGBOOK.md hash-object 1d9f07faa344e000f94985e3b20bd8a58c97b233 before and after; git status shows only the 5 new .research files.
