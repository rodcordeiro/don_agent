param(
    [string]$Version = "dev",
    [ValidateSet("auto", "docker", "go")]
    [string]$Builder = "auto",
    [string]$OutputDir = "dist",
    [switch]$Clean
)

$ErrorActionPreference = "Stop"

function Resolve-RepoRoot {
    $scriptDir = Split-Path -Parent $PSCommandPath
    return (Resolve-Path (Join-Path $scriptDir "..")).Path
}

function Test-Command {
    param([string]$Name)
    $null -ne (Get-Command $Name -ErrorAction SilentlyContinue)
}

function Test-DockerAvailable {
    if (-not (Test-Command "docker")) {
        return $false
    }

    try {
        docker info *> $null
        return $LASTEXITCODE -eq 0
    }
    catch {
        return $false
    }
}

function Assert-NativeCommandSucceeded {
    param([string]$CommandName)

    if ($LASTEXITCODE -ne 0) {
        throw "$CommandName failed with exit code $LASTEXITCODE."
    }
}

function Select-Builder {
    param([string]$Requested)

    if ($Requested -ne "auto") {
        return $Requested
    }

    if (Test-DockerAvailable) {
        return "docker"
    }

    if (Test-Command "go") {
        return "go"
    }

    throw "Neither docker nor go was found. Install Docker or Go, or pass -Builder explicitly."
}

function Invoke-GoBuild {
    param(
        [string]$Builder,
        [string]$RepoRoot,
        [string]$Goos,
        [string]$Goarch,
        [string]$OutputPath
    )

    if ($Builder -eq "docker") {
        docker compose run --no-deps --rm `
            -e "CGO_ENABLED=0" `
            -e "GOOS=$Goos" `
            -e "GOARCH=$Goarch" `
            dev go build -trimpath -ldflags="-s -w" -o "/workspace/$OutputPath" ./cmd/donagent
        Assert-NativeCommandSucceeded "docker compose run"
        return
    }

    $oldCgo = $env:CGO_ENABLED
    $oldGoos = $env:GOOS
    $oldGoarch = $env:GOARCH

    try {
        $env:CGO_ENABLED = "0"
        $env:GOOS = $Goos
        $env:GOARCH = $Goarch
        go build -trimpath -ldflags="-s -w" -o (Join-Path $RepoRoot $OutputPath) ./cmd/donagent
        Assert-NativeCommandSucceeded "go build"
    }
    finally {
        $env:CGO_ENABLED = $oldCgo
        $env:GOOS = $oldGoos
        $env:GOARCH = $oldGoarch
    }
}

function New-ZipArchive {
    param(
        [string]$SourceDir,
        [string]$ArchivePath
    )

    if (Test-Path $ArchivePath) {
        Remove-Item -LiteralPath $ArchivePath -Force
    }

    Compress-Archive -Path (Join-Path $SourceDir "*") -DestinationPath $ArchivePath
}

function New-TarGzArchive {
    param(
        [string]$RepoRoot,
        [string]$SourceDir,
        [string]$ArchivePath
    )

    if (Test-Path $ArchivePath) {
        Remove-Item -LiteralPath $ArchivePath -Force
    }

    if (-not (Test-Command "tar")) {
        throw "tar was not found; cannot create .tar.gz archive."
    }

    $parent = Split-Path -Parent $SourceDir
    $leaf = Split-Path -Leaf $SourceDir
    tar -czf $ArchivePath -C $parent $leaf
    Assert-NativeCommandSucceeded "tar"
}

$repoRoot = Resolve-RepoRoot
Set-Location $repoRoot

$selectedBuilder = Select-Builder $Builder
if ([System.IO.Path]::IsPathRooted($OutputDir)) {
    throw "OutputDir must be a relative directory inside the repository."
}
$resolvedOutputDir = Join-Path $repoRoot $OutputDir

if ($Clean -and (Test-Path $resolvedOutputDir)) {
    $resolved = (Resolve-Path $resolvedOutputDir).Path
    if (-not $resolved.StartsWith($repoRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Refusing to clean output directory outside repository: $resolved"
    }
    Remove-Item -LiteralPath $resolved -Recurse -Force
}

New-Item -ItemType Directory -Force $resolvedOutputDir | Out-Null

$targets = @(
    @{ Goos = "windows"; Goarch = "amd64"; Binary = "donagent.exe"; Archive = "zip" },
    @{ Goos = "linux"; Goarch = "amd64"; Binary = "donagent"; Archive = "tar.gz" }
)

$artifacts = @()

foreach ($target in $targets) {
    $platform = "$($target.Goos)-$($target.Goarch)"
    $packageName = "donagent-$Version-$platform"
    $packageDir = Join-Path (Join-Path $resolvedOutputDir "portable") $packageName
    $relativeBinary = "$OutputDir/portable/$packageName/$($target.Binary)"

    if (Test-Path $packageDir) {
        Remove-Item -LiteralPath $packageDir -Recurse -Force
    }

    New-Item -ItemType Directory -Force (Join-Path $packageDir "assets") | Out-Null

    Invoke-GoBuild `
        -Builder $selectedBuilder `
        -RepoRoot $repoRoot `
        -Goos $target.Goos `
        -Goarch $target.Goarch `
        -OutputPath $relativeBinary

    Copy-Item -LiteralPath (Join-Path $repoRoot "config.example.toml") -Destination (Join-Path $packageDir "config.example.toml")
    Copy-Item -LiteralPath (Join-Path $repoRoot "assets\logo.png") -Destination (Join-Path $packageDir "assets\logo.png")

    if ($target.Archive -eq "zip") {
        $archivePath = Join-Path $resolvedOutputDir "$packageName.zip"
        New-ZipArchive -SourceDir $packageDir -ArchivePath $archivePath
    }
    else {
        $archivePath = Join-Path $resolvedOutputDir "$packageName.tar.gz"
        New-TarGzArchive -RepoRoot $repoRoot -SourceDir $packageDir -ArchivePath $archivePath
    }

    $artifacts += $archivePath
}

Write-Host "Portable artifacts created with builder '$selectedBuilder':"
$artifacts | ForEach-Object { Write-Host " - $_" }
