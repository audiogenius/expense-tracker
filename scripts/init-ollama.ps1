# PowerShell script to initialize Ollama with qwen2.5:0.5b model for 2GB RAM server

Write-Host "🚀 Initializing Ollama for 2GB RAM server..." -ForegroundColor Green

# Wait for Ollama to be ready
Write-Host "⏳ Waiting for Ollama to be ready..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0

while ($attempt -lt $maxAttempts) {
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:11434/api/tags" -Method Get -TimeoutSec 5
        Write-Host "✅ Ollama is ready!" -ForegroundColor Green
        break
    }
    catch {
        Write-Host "⏳ Ollama not ready yet, waiting 10 seconds... (attempt $($attempt + 1)/$maxAttempts)" -ForegroundColor Yellow
        Start-Sleep -Seconds 10
        $attempt++
    }
}

if ($attempt -eq $maxAttempts) {
    Write-Host "❌ Ollama failed to start within 5 minutes" -ForegroundColor Red
    exit 1
}

# Check available models
Write-Host "📋 Checking available models..." -ForegroundColor Cyan
try {
    $models = Invoke-RestMethod -Uri "http://localhost:11434/api/tags" -Method Get
    $models.models | ForEach-Object { Write-Host "  - $($_.name)" -ForegroundColor White }
}
catch {
    Write-Host "⚠️ Could not list models" -ForegroundColor Yellow
}

# Pull the lightweight model
Write-Host "📥 Pulling qwen2.5:0.5b model (this may take a few minutes)..." -ForegroundColor Cyan
$pullBody = @{
    name = "qwen2.5:0.5b"
} | ConvertTo-Json

try {
    Invoke-RestMethod -Uri "http://localhost:11434/api/pull" -Method Post -Body $pullBody -ContentType "application/json"
    Write-Host "✅ Model pull completed!" -ForegroundColor Green
}
catch {
    Write-Host "❌ Failed to pull model: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Wait for model to be available
Write-Host "⏳ Waiting for model to be available..." -ForegroundColor Yellow
$maxAttempts = 20
$attempt = 0

while ($attempt -lt $maxAttempts) {
    try {
        $models = Invoke-RestMethod -Uri "http://localhost:11434/api/tags" -Method Get
        if ($models.models | Where-Object { $_.name -like "*qwen2.5:0.5b*" }) {
            Write-Host "✅ Model qwen2.5:0.5b is ready!" -ForegroundColor Green
            break
        }
    }
    catch {
        # Continue waiting
    }
    
    Write-Host "⏳ Model not ready yet, waiting 15 seconds... (attempt $($attempt + 1)/$maxAttempts)" -ForegroundColor Yellow
    Start-Sleep -Seconds 15
    $attempt++
}

if ($attempt -eq $maxAttempts) {
    Write-Host "❌ Model failed to load within 5 minutes" -ForegroundColor Red
    exit 1
}

# Test the model
Write-Host "🧪 Testing model with a simple prompt..." -ForegroundColor Cyan
$testBody = @{
    model = "qwen2.5:0.5b"
    prompt = "Привет! Как дела?"
    stream = $false
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:11434/api/generate" -Method Post -Body $testBody -ContentType "application/json"
    if ($response.response) {
        Write-Host "✅ Model test successful!" -ForegroundColor Green
        Write-Host "📝 Response: $($response.response)" -ForegroundColor White
    } else {
        Write-Host "❌ Model test failed - no response" -ForegroundColor Red
        exit 1
    }
}
catch {
    Write-Host "❌ Model test failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Check memory usage
Write-Host "💾 Checking memory usage..." -ForegroundColor Cyan
try {
    $memoryUsage = docker stats expense_ollama --no-stream --format "table {{.MemUsage}}" | Select-Object -Last 1
    Write-Host "📊 Ollama memory usage: $memoryUsage" -ForegroundColor White
}
catch {
    Write-Host "⚠️ Could not check memory usage" -ForegroundColor Yellow
}

# List available models
Write-Host "📋 Available models:" -ForegroundColor Cyan
try {
    $models = Invoke-RestMethod -Uri "http://localhost:11434/api/tags" -Method Get
    $models.models | ForEach-Object {
        $sizeGB = [math]::Round($_.size / 1024 / 1024 / 1024, 2)
        Write-Host "  - $($_.name) (${sizeGB}GB)" -ForegroundColor White
    }
}
catch {
    Write-Host "⚠️ Could not list models" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🎉 Ollama initialization completed successfully!" -ForegroundColor Green
Write-Host "🔗 Ollama API: http://localhost:11434" -ForegroundColor Cyan
Write-Host "📊 Analytics Service: http://localhost:8081" -ForegroundColor Cyan
Write-Host "💡 Model: qwen2.5:0.5b (optimized for 2GB RAM)" -ForegroundColor Cyan
