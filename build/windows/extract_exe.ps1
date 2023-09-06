param (
    [string]$INTEGRATION="none",
    [string]$ARCH="amd64",
    [string]$TAG="v0.0.0"
)
write-host "===> Creating dist folder"
New-Item -ItemType directory -Path .\dist -Force

$VERSION=${TAG}.substring(1)
$exe_folder="nri-${INTEGRATION}_windows_${ARCH}"
$zip_name="nri-${INTEGRATION}-${ARCH}.${VERSION}.zip"

write-host "===> Expanding"
expand-archive -path "dist\${zip_name}" -destinationpath "dist\${exe_folder}\"
