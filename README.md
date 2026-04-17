# Minecraft Fabric 1.20.4 One-EXE Launcher

Лаунчер на Go для запуска **Fabric сервера 1.20.4** одним `.exe`.

## Что делает

При запуске `mc-server-launcher.exe`:

- создаёт/читает `launcher.json`;
- выделяет серверу **RAM 4-6 ГБ** (`-Xms4G -Xmx6G` по умолчанию);
- создаёт папку `mc_server` и `mc_server/mods`;
- скачивает Fabric server launcher jar (если его ещё нет);
- записывает `eula=true`;
- создаёт `SERVER_IP.txt` с IP `212.0.213.86`;
- запускает сервер командой `java -jar fabric-server-launch.jar nogui`;
- выставляет прогрузку чанков в `server.properties`: `view-distance=6` и `simulation-distance=6`.

## Как собрать .exe

```powershell
set GOOS=windows
set GOARCH=amd64
go build -o mc-server-launcher.exe .
```

## Как сделать один архив для себя и друзей

1. Соберите `mc-server-launcher.exe`.
2. Создайте папку, например `minecraft-fabric-server`.
3. Положите в неё:
   - `mc-server-launcher.exe`
   - `launcher.json` (опционально, создастся автоматически)
   - `SERVER_IP.txt` (опционально, создастся автоматически)
   - `README.md`
4. Запакуйте папку в `.zip` и отправьте архив.
5. После распаковки запускайте **только** `mc-server-launcher.exe`.


## Готовый архив одной командой

Скрипт `package.ps1` собирает `mc-server-launcher.exe` и формирует архив `dist/minecraft-fabric-server.zip`, куда уже входят:

- `mc-server-launcher.exe`;
- `launcher.json`;
- `SERVER_IP.txt`;
- `README.md`;
- `mc_server/fabric-server-launch.jar` (Fabric 1.20.4 server core);
- `mc_server/eula.txt`;
- `mc_server/server.properties` с дистанцией чанков 6;
- `mc_server/mods` для ваших модов.

Запуск:

```powershell
powershell -ExecutionPolicy Bypass -File .\package.ps1
```

## Конфиг `launcher.json`

```json
{
  "minecraft_version": "1.20.4",
  "min_ram": "4G",
  "max_ram": "6G",
  "server_dir": "mc_server",
  "jar_name": "fabric-server-launch.jar",
  "server_url": "",
  "server_ip": "212.0.213.86"
}
```

- `minecraft_version` — версия Minecraft (по умолчанию `1.20.4`).
- `min_ram`, `max_ram` — память JVM (по умолчанию от 4 ГБ до 6 ГБ).
- `server_dir` — папка сервера.
- `jar_name` — имя Fabric server launcher jar.
- `server_url` — прямой URL для jar (если нужен ручной контроль).
- `server_ip` — IP, который записывается в `SERVER_IP.txt`.

## Добавление модов

Просто кидайте `.jar` модов в папку:

```text
mc_server/mods
```

Можно добавлять моды в любой момент (обычно лучше на остановленном сервере).

## Важно

- Нужна Java 21+ в `PATH`.
- Для входа друзей откройте порт `25565` на роутере (port forwarding) на ваш ПК.
