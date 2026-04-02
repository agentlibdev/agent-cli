[CmdletBinding()]
param(
  [string]$Version = 'latest'
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$ProjectName = 'agentlib'
$Repo = 'agentlibdev/agent-cli'
$InstallDir = Join-Path $env:USERPROFILE '.agentlib'
$InstallDir = Join-Path $InstallDir 'bin'

function Normalize-Version {
  param([string]$InputVersion)

  if ([string]::IsNullOrWhiteSpace($InputVersion) -or $InputVersion -eq 'latest') {
    return 'latest'
  }

  if ($InputVersion.StartsWith('v')) {
    return $InputVersion
  }

  return "v$InputVersion"
}

function Get-Arch {
  try {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
  } catch {
    $arch = $env:PROCESSOR_ARCHITEW6432
    if (-not $arch) {
      $arch = $env:PROCESSOR_ARCHITECTURE
    }
  }

  switch ($arch.ToUpperInvariant()) {
    'X64' { return 'amd64' }
    'AMD64' { return 'amd64' }
    'ARM64' { return 'arm64' }
    default { throw "error: unsupported architecture: $arch" }
  }
}

function Download-File {
  param(
    [string]$Uri,
    [string]$Destination
  )

  try {
    Invoke-WebRequest -Uri $Uri -OutFile $Destination -Headers @{ 'User-Agent' = 'agentlib-installer' } | Out-Null
  } catch {
    throw "error: download failed for $Uri. $($_.Exception.Message)"
  }
}

function Get-LatestReleaseTag {
  try {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -Headers @{ 'User-Agent' = 'agentlib-installer' }
  } catch {
    throw "error: latest release lookup failed. $($_.Exception.Message)"
  }
  if (-not $release.tag_name) {
    throw 'error: could not determine latest release tag'
  }

  return $release.tag_name
}

function Get-ChecksumForFile {
  param(
    [string]$ChecksumFile,
    [string]$ArchiveName
  )

  $line = Get-Content -LiteralPath $ChecksumFile | Where-Object { $_ -match [regex]::Escape($ArchiveName) } | Select-Object -First 1
  if (-not $line) {
    throw "error: could not find checksum for $ArchiveName"
  }

  return ($line -split '\s+')[0]
}

function Normalize-PathEntry {
  param([string]$Value)

  if ([string]::IsNullOrWhiteSpace($Value)) {
    return $null
  }

  $Expanded = [Environment]::ExpandEnvironmentVariables($Value.Trim())
  try {
    return [System.IO.Path]::GetFullPath($Expanded).ToLowerInvariant()
  } catch {
    return $Expanded.ToLowerInvariant()
  }
}

$ReleaseTag = Normalize-Version $Version
if ($ReleaseTag -eq 'latest') {
  $ReleaseTag = Get-LatestReleaseTag
}

$BaseUrl = if ($ReleaseTag -eq 'latest') {
  "https://github.com/$Repo/releases/latest/download"
} else {
  "https://github.com/$Repo/releases/download/$ReleaseTag"
}

$Arch = Get-Arch
$ArchiveName = "${ProjectName}_${ReleaseTag}_windows_${Arch}.zip"
$ChecksumName = "${ProjectName}_checksums.txt"
$BinaryName = "${ProjectName}.exe"
$BinaryPath = Join-Path $InstallDir $BinaryName

$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) "agentlib-install-$([guid]::NewGuid().ToString('N'))"
New-Item -ItemType Directory -Path $TempDir | Out-Null

try {
  $ChecksumPath = Join-Path $TempDir $ChecksumName
  $ArchivePath = Join-Path $TempDir $ArchiveName

  Download-File "$BaseUrl/$ChecksumName" $ChecksumPath
  Download-File "$BaseUrl/$ArchiveName" $ArchivePath

  $ExpectedSha = Get-ChecksumForFile $ChecksumPath $ArchiveName
  $ActualSha = (Get-FileHash -LiteralPath $ArchivePath -Algorithm SHA256).Hash
  if ($ActualSha.ToLowerInvariant() -ne $ExpectedSha.ToLowerInvariant()) {
    throw "error: checksum mismatch for $ArchiveName"
  }

  try {
    Expand-Archive -LiteralPath $ArchivePath -DestinationPath $TempDir -Force
  } catch {
    throw "error: archive extraction failed for $ArchiveName. $($_.Exception.Message)"
  }

  $ExtractedBinary = Get-ChildItem -LiteralPath $TempDir -Recurse -File -Filter $BinaryName | Select-Object -First 1
  if (-not $ExtractedBinary) {
    throw "error: extracted binary not found: $BinaryName"
  }

  New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
  Copy-Item -LiteralPath $ExtractedBinary.FullName -Destination $BinaryPath -Force

  Write-Host "agentlib installed to $BinaryPath"

  $UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
  $UserPathEntries = @()
  if ($UserPath) {
    $UserPathEntries = $UserPath -split ';'
  }
  $NormalizedInstallDir = Normalize-PathEntry $InstallDir
  $NormalizedUserPathEntries = @($UserPathEntries | ForEach-Object { Normalize-PathEntry $_ } | Where-Object { $_ })

  if ($NormalizedUserPathEntries -contains $NormalizedInstallDir) {
    Write-Host "User PATH already includes $InstallDir"
  } else {
    Write-Host "Add this directory to your user PATH, then open a new terminal:"
    Write-Host "  $InstallDir"
    Write-Host "Use the Windows Environment Variables UI or your shell profile to persist it."
  }
} finally {
  Remove-Item -LiteralPath $TempDir -Recurse -Force -ErrorAction SilentlyContinue
}
