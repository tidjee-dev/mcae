# mcae — Minecraft Companion Assets Extractor

CLI to extract and inspect Minecraft mod, modpack and vanilla assets.

## Installation

```bash
go install github.com/tidjee-dev/mcae@latest
```

## Usage

```bash
mcae --help
mcae extract --help

mcae extract mod ./mods/create-6.0.6.jar
mcae extract mod ./mods/create-6.0.6.jar --output ./data

mcae extract modpack ./data/modpacks/cuboid-outpost-luxury
mcae extract modpack ./data/modpacks/cuboid-outpost-luxury --output ./data/extracted

mcae extract vanilla 1.21.8
mcae extract vanilla 1.21.8 --output ./data/minecraft

mcae inspect ./extracted/create-6.0.6
mcae inspect ./mods/create-6.0.6.jar
mcae inspect ./data/modpacks/cuboid-outpost-luxury

mcae version
```

Global flags: `--no-color`, `--quiet/-q`. `NO_COLOR` is also honored.
Output is styled with Lipgloss + Bubbles (spinner/progress/table) on a TTY
and degrades to plain text when piped or with `--no-color`.

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

## Version

Version is centralized in `internal/version`:

```go
const (
    Name    = "mcae"
    Version = "0.1.0"
)
```

Override at build time:

```bash
go build -ldflags "-X github.com/tidjee-dev/mcae/internal/version.Version=x.y.z" ./...
```
