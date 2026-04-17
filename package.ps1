$ErrorActionPreference = "Stop"

$distDir = "dist/minecraft-fabric-server"
$zipPath = "dist/minecraft-fabric-server.zip"
$serverIp = "212.0.213.86"

New-Item -ItemType Directory -Force -Path $distDir | Out-Null

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o "$distDir/mc-server-launcher.exe" .

Copy-Item README.md "$distDir/README.md" -Force

if (Test-Path launcher.json) {
    Copy-Item launcher.json "$distDir/launcher.json" -Force
}
else {
@'
{
  "minecraft_version": "1.20.4",
  "min_ram": "4G",
  "max_ram": "6G",
  "server_dir": "mc_server",
  "jar_name": "fabric-server-launch.jar",
  "server_url": "",
  "server_ip": "212.0.213.86"
}
'@ | Set-Content -Path "$distDir/launcher.json" -Encoding UTF8
}

"IP сервера: $serverIp`nПорт по умолчанию: 25565" | Set-Content -Path "$distDir/SERVER_IP.txt" -Encoding UTF8

if (Test-Path $zipPath) {
    Remove-Item $zipPath -Force
}

Compress-Archive -Path "$distDir/*" -DestinationPath $zipPath
Write-Host "Done: $zipPath"
