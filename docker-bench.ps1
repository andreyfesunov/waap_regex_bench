param(
    [ValidateSet('hyperscan', 'baseline', 'test', 'environment')][string]$Engine = 'hyperscan',
    [ValidateRange(1, 2147483647)][int]$Workers = 1,
    [ValidateRange(0, 2147483647)][int]$Limit = 100000,
    [ValidateSet('all', 'common')][string]$PatternSet = 'all'
)

$ErrorActionPreference = 'Stop'
$dockerArgs = @('run', '--rm', '--platform', 'linux/amd64')

if ($Engine -eq 'test') {
    & docker @dockerArgs waap-regex-bench sh -c 'cd /src && go test ./... && go test -tags hyperscan ./... && go test -race -tags hyperscan ./internal/... && go vet -tags hyperscan ./...'
    exit $LASTEXITCODE
}
if ($Engine -eq 'environment') {
    & docker @dockerArgs waap-regex-bench sh -c 'go version; pkg-config --modversion libhs; nproc; cat /sys/fs/cgroup/cpu.max; cat /sys/fs/cgroup/memory.max; lscpu'
    exit $LASTEXITCODE
}

$repoRoot = $PSScriptRoot
$resultsRoot = Join-Path $repoRoot 'bench-results'
New-Item -ItemType Directory -Path $resultsRoot -Force | Out-Null
$prefix = "linux-$Engine-$PatternSet-$Workers-$Limit"
$dockerArgs += @('--mount', "type=bind,source=$repoRoot\dataset,target=/dataset,readonly", '--mount', "type=bind,source=$resultsRoot,target=/results")
if ($Limit -eq 0 -and $Engine -eq 'hyperscan') { $dockerArgs += @('--env', 'GOMEMLIMIT=6GiB') }
$dockerArgs += 'waap-regex-bench'
$programArgs = @($Engine, '-dataset', '/dataset', '-workers', "$Workers", '-limit', "$Limit", '-output', "/results/$prefix.jsonl")
if ($Engine -eq 'hyperscan') { $programArgs += @('-pattern-set', $PatternSet) }
elseif ($PatternSet -eq 'common') { $programArgs += @('-pattern-ids', "/results/linux-hyperscan-common-$Workers-$Limit.jsonl.patterns.json") }

function Quote-Shell([string]$value) { return "'" + $value.Replace("'", "'\''") + "'" }
$command = ($programArgs | ForEach-Object { Quote-Shell $_ }) -join ' '
$command += " > $(Quote-Shell "/results/$prefix.json") 2> $(Quote-Shell "/results/$prefix.log")"
& docker @dockerArgs sh -c $command
exit $LASTEXITCODE
