<#

.SYNOPSIS
Build a Go application to various platforms and create the folder structure for the go module "updater"

.NOTES
Don't move this script, is must be in the same folder as cli.go.

#>

Param(
    [Parameter(Mandatory = $False)]
    [string]$version = "1.0.0",

    [Parameter(Mandatory = $False)]
    [string]$appName = "myCore",

    [Parameter(Mandatory = $False)]
    [string]$channel = "Beta",

    [Parameter(Mandatory = $True)]
    [string]$rootFolder = "",

    [Parameter(Mandatory = $False)]
    [string]$sign = "true"
)


if ($null -eq (Get-Command Go -ErrorAction Ignore)) {
    Write-Error "Couldn't find Go, is PATH to Go missing?"
    return
}

$buildTime = Get-Date -Format "yyyy-MM-dd_HH:mm:ss"

# Just uncomment the platfoms you don't need
$platforms = @()
$platforms += @{GOOS = "windows"; GOARCH = "amd64"; }
$platforms += @{GOOS = "windows"; GOARCH = "386"; }
#$platforms += @{GOOS = "linux"; GOARCH = "amd64"; }
#$platforms += @{GOOS = "linux"; GOARCH = "386"; }
#$platforms += @{GOOS = "linux"; GOARCH = "arm"; } -
#$platforms += @{GOOS = "linux"; GOARCH = "arm64"; } -
#$platforms += @{GOOS = "darwin"; GOARCH = "amd64"; }

# Extract Major Version
if ($version -match "^([0-9]+).") {
    $major = $matches[1]
}
else {
    $major = "Unknown"
}

if ($sign -eq "true") {
    .\minisign -G
    $pubKey = (Get-Content "minisign.pub")[1]
    Write-Host $pubKey
}

$majorFolder = [System.IO.Path]::Combine($appName, $channel)
$versionFolder = [System.IO.Path]::Combine($majorFolder, $major)
$buildFolder = [System.IO.Path]::Combine($rootFolder, "build" , $versionFolder)

$latestFileName = "latest.txt"

$builds = @()

# Build
foreach ($item in $platforms ) {
    Write-Host "Build" $item.GOOS $item.GOARCH  -ForegroundColor Green

    $env:GOOS = $item.GOOS
    $env:GOARCH = $item.GOARCH

    if ($item.GOOS -eq "windows") {
        $extension = ".exe"
    }
    else {
        $extension = $null
    }
   
    $fileName = ("{0}_{1}_{2}_{3}{4}" -f $appName, $item.GOARCH, $item.GOOS, $version, $extension)
    $buildOutput = [System.IO.Path]::Combine($buildFolder, $fileName)
    if ($sign -eq "false") {
        $executeExpression = "go build -ldflags ""-s -w -X main.AppName={0} -X main.Channel={1} -X main.Architecture={2} -X main.Plattform={3} -X main.Version={4} -X main.BuildTime={5}"" -trimpath -o {6}" -f $appName, $channel, $item.GOARCH, $item.GOOS, $version, $buildTime, $buildOutput
    }
    else {
        $executeExpression = "go build -ldflags ""-s -w -X main.AppName={0} -X main.Channel={1} -X main.Architecture={2} -X main.Plattform={3} -X main.Version={4} -X main.BuildTime={5} -X github.com/haevg-rz/go-updater/updater.UpdateFilesPubKey={6}"" -trimpath -o {7}" -f $appName, $channel, $item.GOARCH, $item.GOOS, $version, $buildTime, $pubKey, $buildOutput
    }
    Write-Host "Execute", $executeExpression -ForegroundColor Gray
    Invoke-Expression $executeExpression

    if ($LASTEXITCODE -ne 0) {
        Write-Host "Abort!"  -ForegroundColor Red
        continue
    }

    if ($sign -eq "true") {
        .\minisign -S -m $buildOutput
    }

    $build = [ordered]@{
        asset         = $appName;
        channel       = $channel;
        version       = $version; 
        specs         = @{plattform = $item.GOOS; architecture = $item.GOARCH };
        fileExtension = $extension;
        filePath      = ([System.IO.Path]::Combine($versionFolder, $fileName));
        buildTime     = $buildTime;
    }
    
    $builds += $build
}

# Write "Latest" files
[System.IO.File]::WriteAllText([System.IO.Path]::Combine($buildFolder, $latestFileName), $version)

$latestMajor = [System.IO.Path]::Combine($rootFolder, "build", $majorFolder, $latestFileName)
$data = Get-Content -erroraction "silentlycontinue" $latestMajor
if (($null -eq $data) -or $major -gt $data) {
    [System.IO.File]::WriteAllText($latestMajor, $major)
}
$buildsJSON = $builds | ConvertTo-Json -Depth 3
[System.IO.File]::WriteAllText([System.IO.Path]::Combine($buildFolder, ("{0}.JSON" -f $version)), $buildsJSON)

Write-Host "Done!"