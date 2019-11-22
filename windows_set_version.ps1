param (
	 [int]$major = $(throw "-major is required"),
	 [int]$minor = $(throw "-minor is required"),
	 [int]$patch = $(throw "-patch is required"),
	 [int]$build = 0
)

$integration = $(Split-Path -Leaf $PSScriptRoot)
$integrationName = $integration.Replace("nri-", "")
$executable = "nri-$integrationName.exe"

if (-not (Test-Path env:GOPATH)) {
	Write-Error "GOPATH not defined."
}
$projectRootPath = Join-Path -Path $env:GOPATH -ChildPath "src\github.com\newrelic\$integration"

$versionInfoTempl = Get-Childitem -Path $projectRootPath -Include "versioninfo.json.template" -Recurse -ErrorAction SilentlyContinue
if ("$versionInfoTempl" -eq "") {
	Write-Error "$versionInfoTempl not found."
	exit 0
}
$versionInfoPath = $versionInfoTempl.DirectoryName + "\versioninfo.json"
Copy-Item -Path $versionInfoTempl -Destination $versionInfoPath -Force

$versionInfo = Get-Content -Path $versionInfoPath -Encoding UTF8
$versionInfo = $versionInfo -replace "{MajorVersion}", $major
$versionInfo = $versionInfo -replace "{MinorVersion}", $minor
$versionInfo = $versionInfo -replace "{PatchVersion}", $patch
$versionInfo = $versionInfo -replace "{BuildVersion}", $build
$versionInfo = $versionInfo -replace "{Integration}", $integration
$versionInfo = $versionInfo -replace "{IntegrationExe}", $executable
$versionInfo = $versionInfo -replace "{Year}", (Get-Date).year
Set-Content -Path $versionInfoPath -Value $versionInfo

$wix386Path = Join-Path -Path $projectRootPath -ChildPath "pkg\windows\nri-386-installer\Product.wxs"
$wixAmd64Path = Join-Path -Path $projectRootPath -ChildPath "pkg\windows\nri-amd64-installer\Product.wxs"

Function ProcessProductFile($productPath) {
	if ((Test-Path "$productPath.template" -PathType Leaf) -eq $False) {
		Write-Error "$productPath.template not found."
	}
	Copy-Item -Path "$productPath.template" -Destination $productPath -Force

	$product = Get-Content -Path $productPath -Encoding UTF8
	$product = $product -replace "{IntegrationVersion}", "$major.$minor.$patch"
	$product = $product -replace "{Year}", (Get-Date).year
	$product = $product -replace "{IntegrationExe}", $executable
	$product = $product -replace "{IntegrationName}", $integrationName
	Set-Content -Value $product -Path $productPath
}

ProcessProductFile($wix386Path)
ProcessProductFile($wixAmd64Path)
