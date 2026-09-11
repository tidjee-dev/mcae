# mcae — Minecraft Companion Assets Extractor

[![Go Reference](https://pkg.go.dev/badge/github.com/tidjee-dev/mcae.svg)](https://pkg.go.dev/github.com/tidjee-dev/mcae)

CLI to extract and inspect Minecraft mod, modpack and vanilla assets.

> **Disclaimer:** `mcae` is an unofficial, community-made tool and is not
> affiliated with, endorsed by, or connected to Mojang Studios or
> Microsoft.

## Requirements

- Go `1.27.1` (see `go.mod`)
- `gh` CLI (optional, for the PR workflow)

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

mcae extract mod <mod_jar_file>
mcae extract mod <mod_jar_file> --output <output_dir>
mcae extract mod <mod_jar_file> --force

mcae extract modpack <modpack_dir>
mcae extract modpack <modpack_dir> --output <output_dir>
mcae extract modpack <modpack_dir> --force

mcae extract vanilla <minecraft_version>
mcae extract vanilla <minecraft_version> --output <output_dir>
mcae extract vanilla <minecraft_version> --force

mcae inspect <extracted_dir>
mcae inspect <mod_jar_file>
mcae inspect <modpack_dir>

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
mcae extract mod <mod_jar_file> -q
# ./extracted/<mod_name>
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
./extracted/<mod_name>/
# or with --output <output_dir>:
<output_dir>/
└── <mod_name>/
```

Modpack (each JAR gets its own directory):

```text
<modpack_dir>/
├── mods/
│   ├── <mod_name>.jar
│   └── <other_mod>.jar
└── extracted/
    ├── <mod_name>/
    └── <other_mod>/
```

Vanilla (`--output` is a base directory):

```text
./extracted/vanilla/<minecraft_version>/
# with --output <output_dir>:
<output_dir>/vanilla/<minecraft_version>/
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

API reference is published on
[pkg.go.dev](https://pkg.go.dev/github.com/tidjee-dev/mcae). Note that most
logic lives under `internal/` (CLI implementation detail, not importable);
the module page is the entry point.

## Version

Version is centralized in `internal/version`:

```go
var (
    Name    = "mcae"
    Version = "1.0.0"
)
```

The variables (not constants) allow build-time injection:

```bash
go build -ldflags "-X github.com/tidjee-dev/mcae/internal/version.Version=x.y.z" ./...
```

`mcae version` prints `mcae 1.0.0`; the root `--version` flag prints
`mcae version 1.0.0` (Cobra default template).

## Disclaimer

`mcae` is an unofficial, community-made tool and is not affiliated with,
endorsed by, or connected to Mojang Studios or Microsoft. Minecraft is a
trademark of Mojang Synergies AB. Vanilla assets are downloaded from
Mojang's official servers at runtime and are never bundled or
redistributed by this project.
