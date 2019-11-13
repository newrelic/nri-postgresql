<#
    .SYNOPSIS
        This script verifies, tests, builds and packages a New Relic Infrastructure Integration
#>
param (
    # Target architecture: amd64 (default) or 386
    [ValidateSet("amd64", "386")]
    [string]$arch="amd64",
    [string]$version="0.0.0",
    # Creates a signed installer
    [switch]$installer=$false,
    # Skip tests
    [switch]$skipTests=$false
)

$integration = $(Split-Path -Leaf $PSScriptRoot)
$integrationName = $integration.Replace("nri-", "")
$executable = "nr-$integrationName.exe"

# verifying version number format
$v = $version.Split(".")

if ($v.Length -ne 3) {
    echo "-version must follow a numeric major.minor.patch semantic versioning schema (received: $version)"
    exit -1
}

$wrong = $v | ? { (-Not [System.Int32]::TryParse($_, [ref]0)) -or ( $_.Length -eq 0) -or ([int]$_ -lt 0)} | % { 1 }
if ($wrong.Length  -ne 0) {
    echo "-version major, minor and patch must be valid positive integers (received: $version)"
    exit -1
}

echo "--- Configuring version $version for artifacts"

.\windows_set_version.ps1 -major $v[0] -minor $v[1] -patch $v[2]

echo "--- Checking dependencies"

echo "Checking Go..."
go version
if (-not $?)
{
    echo "Can't find Go"
    exit -1
}

echo "Checking MSBuild.exe..."
$msBuild = (Get-ItemProperty hklm:\software\Microsoft\MSBuild\ToolsVersions\4.0).MSBuildToolsPath
if ($msBuild.Length -eq 0) {
    echo "Can't find MSBuild tool. .NET Framework 4.0.x must be installed"
    exit -1
}
echo $msBuild

$env:GOOS="windows"
$env:GOARCH=$arch

echo "--- Collecting files"

$goFiles = go list ./...

echo "--- Format check"

$wrongFormat = go fmt $goFiles

if ($wrongFormat -and ($wrongFormat.Length -gt 0))
{
    echo "ERROR: Wrong format for files:"
    echo $wrongFormat
    exit -1
}

if (-Not $skipTests) {
    echo "--- Running tests"

    go test $goFiles
    if (-not $?)
    {
        echo "Failed running tests"
        exit -1
    }    
}

echo "--- Running Build"

go build -v $goFiles
if (-not $?)
{
    echo "Failed building files"
    exit -1
}

echo "--- Collecting Go main files"

$packages = go list -f "{{.ImportPath}} {{.Name}}" ./...  | ConvertFrom-String -PropertyNames Path, Name
$mainPackage = $packages | ? { $_.Name -eq 'main' } | % { $_.Path }

echo "generating $integrationName"
go generate $mainPackage

$fileName = ([io.fileinfo]$mainPackage).BaseName

echo "creating $executable"
go build -ldflags "-X main.buildVersion=$version" -o ".\target\bin\windows_$arch\$executable" $mainPackage

If (-Not $installer) {
    exit 0
}

echo "--- Building Installer"

Push-Location -Path "pkg\windows\nri-$arch-installer"
$env:integration = $integration
. $msBuild/MSBuild.exe nri-installer.wixproj

if (-not $?)
{
    echo "Failed building installer"
    Pop-Location
    exit -1
}

echo "Making versioned installed copy"

cd bin\Release

cp "$integration-$arch.msi" "$integration-$arch.$version.msi"
cp "$integration-$arch.msi" "$integration.msi"

Pop-Location
