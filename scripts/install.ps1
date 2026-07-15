#Requires -Version 5.1
<#
.SYNOPSIS
  Install paycrest-mcp into a fixed location and (optionally) wire Cursor mcp.json.

.DESCRIPTION
  Downloads a release binary into:

    %USERPROFILE%\.paycrest\paycrest-mcp.exe

  Prefers the GitHub CLI (`gh`) when available (works with private repos if you ran `gh auth login`).
  Falls back to the GitHub API / direct URL with GH_TOKEN / GITHUB_TOKEN.

.EXAMPLE
  .\scripts\install.ps1

.EXAMPLE
  .\scripts\install.ps1 -Tag v0.1.0 -ApiKey "your-sender-key"
#>
[CmdletBinding()]
param(
  [string]$Tag = "latest",
  [string]$Repo = "paycrest/sender-mcp",
  [switch]$SkipMcpJson,
  [string]$ApiKey = $env:PAYCREST_API_KEY
)

$ErrorActionPreference = "Stop"

$installDir = Join-Path $env:USERPROFILE ".paycrest"
$destExe = Join-Path $installDir "paycrest-mcp.exe"
$mcpPath = Join-Path $env:USERPROFILE ".cursor\mcp.json"
$ghToken = $env:GH_TOKEN
if ([string]::IsNullOrWhiteSpace($ghToken)) {
  $ghToken = $env:GITHUB_TOKEN
}

function Get-GitHubHeaders {
  $headers = @{
    "User-Agent" = "paycrest-mcp-install"
    "Accept"     = "application/vnd.github+json"
  }
  if (-not [string]::IsNullOrWhiteSpace($ghToken)) {
    $headers["Authorization"] = "Bearer $ghToken"
  }
  return $headers
}

function Install-WithGh {
  param([string]$TagName, [string]$RepoName, [string]$OutFile)
  $gh = Get-Command gh -ErrorAction SilentlyContinue
  if (-not $gh) { return $false }

  New-Item -ItemType Directory -Force -Path (Split-Path $OutFile -Parent) | Out-Null
  $tmp = Join-Path $env:TEMP ("paycrest-mcp-dl-" + [guid]::NewGuid().ToString("n"))
  New-Item -ItemType Directory -Force -Path $tmp | Out-Null
  try {
    $args = @("release", "download", "-R", $RepoName, "-p", "*windows_amd64.exe", "-D", $tmp, "--clobber")
    if ($TagName -ne "latest") {
      $args = @("release", "download", $TagName, "-R", $RepoName, "-p", "*windows_amd64.exe", "-D", $tmp, "--clobber")
    }
    Write-Host "Downloading with gh $($args -join ' ') ..."
    & gh @args
    if ($LASTEXITCODE -ne 0) { return $false }
    $found = Get-ChildItem -Path $tmp -Filter "*windows_amd64.exe" | Select-Object -First 1
    if (-not $found) { return $false }
    Copy-Item -Force -Path $found.FullName -Destination $OutFile
    Unblock-File -Path $OutFile -ErrorAction SilentlyContinue
    if ($TagName -eq "latest") {
      $tagOut = & gh release view -R $RepoName --json tagName -q .tagName 2>$null
      if ($tagOut) { $script:ResolvedTag = $tagOut.Trim() }
    } else {
      $script:ResolvedTag = $TagName
    }
    return $true
  } finally {
    Remove-Item -Recurse -Force -Path $tmp -ErrorAction SilentlyContinue
  }
}

function Install-WithApiOrUrl {
  param([string]$TagName, [string]$RepoName, [string]$OutFile)

  $tag = $TagName
  $url = $null
  $name = $null

  try {
    if ($TagName -eq "latest") {
      $uri = "https://api.github.com/repos/$RepoName/releases/latest"
    } else {
      $uri = "https://api.github.com/repos/$RepoName/releases/tags/$TagName"
    }
    $release = Invoke-RestMethod -Uri $uri -Headers (Get-GitHubHeaders)
    $tag = $release.tag_name
    $asset = $release.assets | Where-Object { $_.name -match 'windows_amd64\.exe$' } | Select-Object -First 1
    if (-not $asset) { throw "No windows_amd64.exe on release $tag" }
    $url = $asset.browser_download_url
    $name = $asset.name
  } catch {
    if ($TagName -eq "latest") {
      throw @"
Could not download release ($($_.Exception.Message)).

Private repo? Use one of:
  1) Install GitHub CLI and login:  gh auth login
     then re-run:  .\scripts\install.ps1
  2) Set a PAT:  `$env:GH_TOKEN = 'ghp_...'
  3) Or download the .exe from the Releases page and copy it to:
     $OutFile
"@
    }
    $name = "paycrest-mcp_${TagName}_windows_amd64.exe"
    $url = "https://github.com/$RepoName/releases/download/$TagName/$name"
    $tag = $TagName
    Write-Host "API unavailable; trying direct URL: $url"
  }

  Write-Host "Downloading $name -> $OutFile"
  New-Item -ItemType Directory -Force -Path (Split-Path $OutFile -Parent) | Out-Null
  Invoke-WebRequest -Uri $url -OutFile $OutFile -UseBasicParsing -Headers (Get-GitHubHeaders)
  Unblock-File -Path $OutFile -ErrorAction SilentlyContinue
  $script:ResolvedTag = $tag
}

$script:ResolvedTag = $Tag
Write-Host "Installing from github.com/$Repo ($Tag) ..."
if (-not (Install-WithGh -TagName $Tag -RepoName $Repo -OutFile $destExe)) {
  Install-WithApiOrUrl -TagName $Tag -RepoName $Repo -OutFile $destExe
}
Write-Host "Installed: $destExe ($ResolvedTag)"

if (-not $SkipMcpJson) {
  $cursorDir = Split-Path $mcpPath -Parent
  New-Item -ItemType Directory -Force -Path $cursorDir | Out-Null

  $commandPath = ($destExe -replace '\\', '/')
  $keyValue = if ([string]::IsNullOrWhiteSpace($ApiKey)) {
    '${env:PAYCREST_API_KEY}'
  } else {
    $ApiKey.Trim()
  }

  $paycrestServer = [pscustomobject]@{
    command = $commandPath
    env     = [pscustomobject]@{
      PAYCREST_API_KEY = $keyValue
    }
  }

  if (Test-Path $mcpPath) {
    $raw = Get-Content -Raw -Path $mcpPath
    if ([string]::IsNullOrWhiteSpace($raw)) {
      $root = [pscustomobject]@{ mcpServers = [pscustomobject]@{ paycrest = $paycrestServer } }
    } else {
      try {
        $root = $raw | ConvertFrom-Json
      } catch {
        throw "Could not parse existing $mcpPath - fix or remove it, then re-run."
      }
      if ($null -eq $root.mcpServers) {
        $root | Add-Member -NotePropertyName mcpServers -NotePropertyValue ([pscustomobject]@{}) -Force
      }
      $servers = $root.mcpServers
      if ($servers.PSObject.Properties.Name -contains 'paycrest') {
        $servers.PSObject.Properties.Remove('paycrest')
      }
      $servers | Add-Member -NotePropertyName paycrest -NotePropertyValue $paycrestServer -Force
    }
  } else {
    $root = [pscustomobject]@{
      mcpServers = [pscustomobject]@{
        paycrest = $paycrestServer
      }
    }
  }

  ($root | ConvertTo-Json -Depth 30) | Set-Content -Path $mcpPath -Encoding utf8
  Write-Host "Updated Cursor MCP config: $mcpPath"
}

Write-Host ""
Write-Host "Next:"
if ([string]::IsNullOrWhiteSpace($ApiKey) -or $ApiKey -eq '${env:PAYCREST_API_KEY}') {
  Write-Host "  1. Set your sender API key (User env or mcp.json):"
  Write-Host '     PAYCREST_API_KEY=<your-dashboard-key>'
  Write-Host "     Or edit $mcpPath and replace `${env:PAYCREST_API_KEY}."
} else {
  Write-Host "  1. API key written into mcp.json (do not commit that file)."
}
Write-Host "  2. Reload MCP in Cursor (Settings -> Tools & MCP), or restart Cursor."
Write-Host "  3. Test: ask the agent to list Paycrest currencies."
Write-Host ""
Write-Host "Fixed binary path: $destExe"
