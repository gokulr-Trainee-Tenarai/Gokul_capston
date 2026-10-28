# PowerShell script to register services directly with Consul HTTP API (Lab 4 pattern)
$services = @(
    @{
        ID      = "product-service-1"
        Name    = "product-service"
        Address = "localhost"
        Port    = 8081
        Check   = @{
            HTTP     = "http://localhost:8081/health/live"
            Interval = "10s"
            Timeout  = "5s"
        }
    },
    @{
        ID      = "order-service-1"
        Name    = "order-service"
        Address = "localhost"
        Port    = 8082
        Check   = @{
            HTTP     = "http://localhost:8082/health/live"
            Interval = "10s"
            Timeout  = "5s"
        }
    }
)

foreach ($s in $services) {
    $body = $s | ConvertTo-Json -Depth 5
    try {
        Invoke-RestMethod -Method Put -Uri "http://localhost:8500/v1/agent/service/register" -ContentType "application/json" -Body $body
        Write-Host "Successfully registered $($s.Name) ($($s.ID)) with Consul." -ForegroundColor Green
    } catch {
        Write-Host "Failed to register $($s.Name): $_" -ForegroundColor Red
    }
}

Write-Host "`nCheck registered services at http://localhost:8500" -ForegroundColor Cyan
