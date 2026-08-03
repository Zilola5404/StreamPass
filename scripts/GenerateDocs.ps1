# StreamPass Documentation Generator
# Validates documentation structure exists and reports status.

$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)

$RequiredDocs = @(
    "docs\00_ProjectRules.md",
    "docs\14_AIContext.md",
    "docs\03_CurrentState.md",
    "docs\08_API.md",
    "docs\07_Architecture.md"
)

$RequiredAI = @(
    "ai\CurrentTask.md",
    "ai\LastSession.md",
    "ai\OpenQuestions.md"
)

Write-Host "=== StreamPass Documentation Check ===" -ForegroundColor Cyan

$Missing = @()
foreach ($file in ($RequiredDocs + $RequiredAI)) {
    $path = Join-Path $Root $file
    if (Test-Path $path) {
        $size = (Get-Item $path).Length
        $status = if ($size -gt 100) { "OK" } else { "EMPTY" }
        Write-Host "[$status] $file ($size bytes)" -ForegroundColor $(if ($status -eq "OK") { "Green" } else { "Yellow" })
        if ($status -eq "EMPTY") { $Missing += $file }
    } else {
        Write-Host "[MISSING] $file" -ForegroundColor Red
        $Missing += $file
    }
}

$DocCount = (Get-ChildItem "$Root\docs\*.md").Count
$AICount = (Get-ChildItem "$Root\ai\*.md").Count
$ReportCount = (Get-ChildItem "$Root\reports\*.md").Count
$PromptCount = (Get-ChildItem "$Root\prompts\*.md").Count

Write-Host "`nCounts: docs=$DocCount, ai=$AICount, reports=$ReportCount, prompts=$PromptCount"

if ($Missing.Count -gt 0) {
    Write-Host "`nMissing or empty: $($Missing -join ', ')" -ForegroundColor Red
    exit 1
}

Write-Host "`nDocumentation structure: OK" -ForegroundColor Green
