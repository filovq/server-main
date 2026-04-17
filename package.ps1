$ErrorActionPreference = "Stop"

$distRoot = "dist"
$distDir = "$distRoot/minecraft-fabric-server"
$zipPath = "$distRoot/minecraft-fabric-server.zip"
$serverDir = "$distDir/mc_server"
$modsDir = "$serverDir/mods"
$serverIp = "212.0.213.86"
$minecraftVersion = "1.20.4"
$jarName = "fabric-server-launch.jar"
$jarPath = "$serverDir/$jarName"

function Get-LatestStableVersion {
    param (
        [Parameter(Mandatory = $true)][string]$Url
    )

    $versions = Invoke-RestMethod -Uri $Url
    if (-not $versions -or $versions.Count -eq 0) {
        throw "Не удалось получить версии из $Url"
    }

    $stable = $versions | Where-Object { $_.stable -eq $true } | Select-Object -First 1
    if ($stable) {
        return $stable.version
    }

    return $versions[0].version
}

New-Item -ItemType Directory -Force -Path $distRoot | Out-Null
New-Item -ItemType Directory -Force -Path $distDir | Out-Null
New-Item -ItemType Directory -Force -Path $serverDir | Out-Null
New-Item -ItemType Directory -Force -Path $modsDir | Out-Null

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

$loaderVersion = Get-LatestStableVersion -Url "https://meta.fabricmc.net/v2/versions/loader"
$installerVersion = Get-LatestStableVersion -Url "https://meta.fabricmc.net/v2/versions/installer"
$jarUrl = "https://meta.fabricmc.net/v2/versions/loader/$minecraftVersion/$loaderVersion/$installerVersion/server/jar"

Write-Host "Downloading Fabric server core from: $jarUrl"
Invoke-WebRequest -Uri $jarUrl -OutFile $jarPath

"eula=true" | Set-Content -Path "$serverDir/eula.txt" -Encoding UTF8
@'
view-distance=6
simulation-distance=6
'@ | Set-Content -Path "$serverDir/server.properties" -Encoding UTF8

"IP сервера: $serverIp`nПорт по умолчанию: 25565" | Set-Content -Path "$distDir/SERVER_IP.txt" -Encoding UTF8

if (Test-Path $zipPath) {
    Remove-Item $zipPath -Force
}

Compress-Archive -Path "$distDir/*" -DestinationPath $zipPath
Write-Host "Done: $zipPath"
