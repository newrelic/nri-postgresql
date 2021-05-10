echo "--- Running tests"

go test ./src/...
if (-not $?)
{
    echo "Failed running tests"
    exit -1
}