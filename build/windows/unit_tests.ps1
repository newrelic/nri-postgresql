echo "--- Running tests"

go test ./...
if (-not $?)
{
    echo "Failed running tests"
    exit -1
}