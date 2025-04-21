$ErrorActionPreference = 'Stop'
$packageName = 'envi'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

# Clean up files
try {
    # Remove the binary
    Remove-Item "$toolsDir\envi.exe" -Force -ErrorAction SilentlyContinue
    
    Write-Host "$packageName has been uninstalled."
} catch {
    Write-Warning "Could not completely remove $packageName. Error: $_"
} 