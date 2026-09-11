# mcae — Minecraft Companion Assets Extractor

CLI to extract and inspect Minecraft mod, modpack and vanilla assets.

## Requirements

* Go `1.27.1` (see `go.mod`)
* `gh` CLI (optional, for the PR workflow)

## Installation

```bash
go install github.com/tidjee-dev/mcae@latest
```

Build from source:

```bash
git clone https://github.com/tidjee-dev/mcae.git
cd mcae
go build ./...
```

## Usage

```bash
mcae --help
mcae extract --help

mcae extract mod ./mods/create-6.0.6.jar
mcae extract mod ./mods/create-6.0.6.jar --output ./data
mcae extract mod ./mods/create-6.0.6.jar --force

mcae extract modpack ./data/modpacks/cuboid-outpost-luxury
mcae extract modpack ./data/modpacks/cuboid-outpost-luxury --output ./data/extracted
mcae extract modpack ./data/modpacks/cuboid-outpost-luxury --force

mcae extract vanilla 1.21.8
mcae extract vanilla 1.21.8 --output ./data/minecraft
mcae extract vanilla 1.21.8 --force

mcae inspect ./extracted/create-6.0.6
mcae inspect ./mods/create-6.0.6.jar
mcae inspect ./data/modpacks/cuboid-outpost-luxury

mcae version
mcae --version
```

Shell completions (Cobra default):

```bash
mcae completion bash|zsh|fish|powershell
```

Global flags: `--no-color`, `--quiet/-q`. `NO_COLOR` is also honored.
Extract commands accept `--force/-f` to remove the destination directory
before extracting; by default existing directories are reused and merged
(a short reuse note is printed unless `--quiet`).

## Output behavior

Output is styled with Lipgloss + Bubbles. On an interactive terminal,
`extract modpack` and `extract vanilla` show a live spinner/progress view
(press `q` to quit); when piped, with `--no-color`, or with `--quiet`,
output falls back to stable streaming lines. `--quiet` prints only the
resulting directories (one per line), which is convenient for scripts:

```bash
mcae extract mod ./mods/example.jar -q
# ./extracted/example
```

Errors are printed once as `Error: …` to stderr without usage spam.
`inspect` prints one summary table with totals plus the namespace list;
modpack inspects show a per-mod table with totals.

Example `inspect` (single mod):

```text
Minecraft Asset Summary

Path: /tmp/mods/example.jar

Namespaces: 1
Section  Category     Files
assets   lang         1
assets   models/item  1
data     recipes      2
Total: 2 assets, 2 data files

Namespaces
  - example
```

Example `extract modpack` (piped/plain):

```text
Found 2 mod JARs
[1/2] example.jar [==========          ]
  assets: 2 files
  data:   2 files
Done: ./extracted/example
[2/2] second.jar [====================]
  assets: 2 files
  data:   2 files
Done: ./extracted/second
Extracted 2 mods: 4 assets, 4 data files in 25ms
```

## Extracted layout

Mod:

```text
./extracted/create-6.0.6/
# or with --output ./data:
data/
└── create-6.0.6/
```

Modpack (each JAR gets its own directory):

```text
cuboid-outpost-luxury/
├── mods/
│   ├── create-6.0.6.jar
│   └── jei-*.jar
└── extracted/
    ├── create-6.0.6/
    └── jei-*/
```

Vanilla (`--output` is a base directory):

```text
./extracted/vanilla/1.21.8/
# with --output ./data/minecraft:
./data/minecraft/vanilla/1.21.8/
```

Vanilla assets are resolved via Mojang's version manifest, the client JAR
is downloaded to a temp file (never committed), then filtered like mods.

## Retained resources

Only these paths are extracted; the namespace stays dynamic
(`minecraft`, `create`, `jei`, … are never hard-coded):

```text
assets/<namespace>/
├── lang/
├── models/
│   ├── block/
│   └── item/
└── textures/
    ├── block/
    ├── entity/
    └── item/

data/<namespace>/
├── loot_tables/
├── recipes/
└── tags/
    ├── blocks/
    └── items/
```

Ignored, for example:

```text
META-INF/
*.class
assets/<namespace>/sounds/
assets/<namespace>/shaders/
assets/<namespace>/blockstates/
data/<namespace>/advancements/
data/<namespace>/functions/
data/<namespace>/structures/
```

JAR entries are Zip-Slip checked; `../../file` style paths are rejected.

## Errors

Common failures (exit 1, stderr only):

```text
Error: mod file does not exist: ./mods/missing.jar
Error: invalid modpack: mods directory not found: ./pack/mods
Error: no mod JARs found in ./pack/mods
Error: minecraft version not found: 9.9.9
```

Argument errors behave the same way; run with `--help` for usage.

## Development

```bash
go fmt ./...
go vet ./...
go test ./...
go build ./...
```

Tests use temp dirs and `archive/zip` fixtures (plus `httptest` for the
vanilla manifest); no real Minecraft or mod JARs are required.

## Version

Version is centralized in `internal/version`:

```go
var (
    Name    = "mcae"
    Version = "0.1.0"
)
```

The variables (not constants) allow build-time injection:

```bash
go build -ldflags "-X github.com/tidjee-dev/mcae/internal/version.Version=x.y.z" ./...
```

`mcae version` prints `mcae 0.1.0`; the root `--version` flag prints
`mcae version 0.1.0` (Cobra default template).
