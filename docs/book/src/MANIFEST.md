# Figure provenance

All figures are original to this book. Each chapter render has a matching
PlantUML source here. The SHA-256 binds the reviewed source; update it when
editing the source and rerender the PNG. No remote includes are needed.

From the repository root, with PlantUML and Graphviz available:

```sh
PLANTUML_LIMIT_SIZE=16384 plantuml -charset UTF-8 -tpng -o "$PWD/docs/book/diagrams" docs/book/src/*.puml
docs/book/check-figures.sh
```

| Stem | Origin | Format | SHA-256 |
| --- | --- | --- | --- |
| context-order | Book original | PlantUML | 8e0e53e399d7830ff41676e812190fd43c3e937f2987d92e7458d2e1888f2fbe |
| launch-route | Book original | PlantUML | e70892310c8d176034d1313b93d3d669b242fd12f2eea869938633eea5d2f24c |
| profile-lock-home | Book original | PlantUML | 6e3d0e9d3ea59570940e85f15a582c7b67cae2fd65fbd18fe3eb3c5f8a5f0195 |
| project-admission | Book original | PlantUML | 3abaa3e4aaa94857e0f637dc94716d0455b5205b5ac06cb2e3c6baac9feae57b |
| werkstatt | Book original | PlantUML | a6166211ea08de6a52d4a3321563dda185df1055deb0b51a26c19f344f2a8ed5 |
| launch-routes-comparison | Book original | PlantUML | 42d1e29c46c2a825627ddbd4ba4b99339af94d064f19eb89833368dd61d7ad5c |
| mcp-delivery | Book original | PlantUML | 7f607ea3bcad3e4f02ba7a7c6f55ea7bd16c0984f320728e51d6d4d59f0914f7 |
| runtime-exclusion | Book original | PlantUML | 51a58894c1cae2330fc1f904c2afe2d008433afa5fbd12050e233fda94771455 |
| trust-boundaries | Book original | PlantUML | 00b8eba88056164ac30dde66a155bba473cde75c5cb847b10cd1ea81841ed5e5 |
