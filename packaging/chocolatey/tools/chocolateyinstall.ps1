$ErrorActionPreference = 'Stop'
$packageName = 'envi'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url64 = 'https://github.com/dexterity-inc/envi/releases/download/v1.0.0/envi-Windows-x86_64.zip'
$checksum64 = '<SHA256_HASH>'

Install-ChocolateyZipPackage $packageName $url64 $toolsDir -checksum64 $checksum64 -checksumType64 'sha256'

# Create shell completions
$binPath = Join-Path $toolsDir "envi.exe"

# Add to PATH if needed
$envPath = $env:PATH
if (-not $envPath.Contains($toolsDir)) {
    Write-Host "Adding $toolsDir to PATH..."
    Install-ChocolateyPath -PathToInstall $toolsDir -PathType 'Machine'
} 