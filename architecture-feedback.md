# Architecture and Code Review Report

## Executive Summary

This report provides a comprehensive review of the Maglev (OneBusAway) Go webapp codebase, focusing on architectural soundness, performance, security, and code quality. The application demonstrates good architectural patterns with clean separation of concerns, but has several critical issues that should be addressed before production deployment.

## Architecture Analysis

### Strengths

1. **Clean Architecture**: Well-separated layers with clear responsibilities
2. **Dependency Injection**: Central `Application` struct holds shared dependencies
3. **Type Safety**: Uses sqlc for compile-time SQL validation
4. **Concurrency**: Proper use of mutexes for real-time data access
5. **Testing Infrastructure**: Good test coverage patterns with test files co-located with code
6. **API Design**: Consistent REST API pattern with middleware for API key validation
7. **Database Abstraction**: Clean separation between in-memory GTFS data and database queries

### Architecture Overview

The application follows a layered approach:
- **Presentation Layer**: REST API handlers and Web UI
- **Application Layer**: Core application logic and dependency container
- **Domain Layer**: Models representing transit data
- **Data Layer**: GTFS manager for both static and real-time data, SQLite database

### Areas for Improvement

1. **Empty Migrations Directory**: Database migrations setup exists but unused
2. **Mixed Data Access**: Both in-memory GTFS data and database queries could lead to confusion
3. **Hardcoded Defaults**: Some default URLs are hardcoded in flags
4. **Limited Error Handling**: Error handling could be more robust
5. **No Graceful Shutdown**: Server doesn't handle shutdown signals properly

## Security Analysis

### Critical Security Issues

1. **Missing HTTP Security Headers**
   - No CORS configuration
   - Missing headers: `X-Content-Type-Options`, `X-Frame-Options`, `Strict-Transport-Security`
   - No Content Security Policy

2. **Weak Secrets Management**
   - API keys passed via command line (visible in process lists)
   - Default API key is "test"
   - No secret rotation mechanism

3. **Missing Rate Limiting**
   - No protection against brute force attacks
   - No request throttling

### Security Strengths

- SQL injection protection via parameterized queries (sqlc)
- Basic API key authentication on all endpoints
- Proper timeout configuration

## Performance Analysis

### Critical Performance Issues

1. **N+1 Query Problems**
   - `routes_for_location_handler.go`: Queries routes for each stop individually
   - `stops_for_location_handler.go`: Multiple queries per stop
   - `schedule_for_stop_handler.go`: Individual queries for each schedule row

2. **Unused Spatial Index**
   - Linear scan through all stops instead of using R-tree index
   - `stops_rtree` table exists but never queried

3. **No Caching Layer**
   - Every request hits database directly
   - No caching for static data or computed results

4. **Inefficient Real-time Updates**
   - Sequential HTTP calls instead of parallel
   - Full data replacement instead of delta updates

5. **Missing Database Optimizations**
   - No connection pooling configuration
   - Row-by-row inserts for bulk data

## Code Quality Issues

### Critical Issues

1. **Fatal Error Handling**
   - Extensive use of `log.Fatal()` crashes the application
   - No graceful error recovery

2. **Resource Leaks**
   - Goroutines run forever with no shutdown mechanism
   - Database connections never explicitly closed
   - No graceful shutdown implementation

3. **Concurrency Bugs**
   - Static GTFS data updated without mutex protection
   - Potential race conditions

### Testing Gaps

- 0% coverage: `cmd/api`, `internal/appconf`, `internal/webui`
- 3.8% coverage: `internal/utils`
- 32.8% coverage: `internal/gtfs`

## Top Recommendations for Junior Engineers/LLMs

### 🚨 Critical Issues (Fix Immediately)

#### 1. **Remove All `log.Fatal()` Calls** ✅ COMPLETED
**Files**: `gtfsdb/helpers.go`, `gtfsdb/client.go`
**Fix**: Replace with error returns and proper error handling
```go
// BAD: log.Fatal("Unable to create DB", err)
// GOOD: return nil, fmt.Errorf("unable to create DB: %w", err)
```
**Status**: ✅ Implemented in commit `41417ae`. All log.Fatal() calls replaced with proper error returns. NewClient() now returns error instead of crashing. Comprehensive tests added for error scenarios.

#### 2. **Add Graceful Shutdown** ✅ COMPLETED
**File**: `cmd/api/main.go`
**Fix**: Implement signal handling for clean shutdown
```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
// Use ctx for server shutdown
```
**Status**: ✅ Implemented in commit `c129128`. Added signal handling for SIGTERM/SIGINT with 30s timeout. Background goroutines properly shut down via shutdown channels and sync.WaitGroup. Resources cleaned up gracefully.

#### 3. **Add Security Headers Middleware** ✅ COMPLETED
**File**: Create `internal/rest_api/security_middleware.go`
**Fix**: Add CORS and security headers to all responses
```go
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        // Add other headers
        next.ServeHTTP(w, r)
    })
}
```
**Status**: ✅ Implemented in commit `edb32d0`. Added comprehensive security headers (HSTS, CSP, XSS protection, etc.) and CORS support. Middleware applied globally to all routes. Includes proper OPTIONS preflight handling.

### ⚡ Performance Issues (High Priority)

#### 4. **Fix N+1 Query Problems** ✅ COMPLETED
**Files**: `internal/rest_api/routes_for_location_handler.go:41`, `internal/rest_api/stops_for_location_handler.go:33,49`
**Fix**: Create batch queries in `gtfsdb/query.sql`
```sql
-- name: GetRoutesForStops :many
SELECT DISTINCT r.* FROM routes r
JOIN stop_times st ON r.route_id = st.route_id
WHERE st.stop_id = ANY($1::text[]);
```
**Status**: ✅ Implemented in commit `4643b44`. Added GetRoutesForStops, GetRouteIDsForStops, and GetAgenciesForStops batch queries. Eliminated N+1 queries in location handlers. Performance improved from O(n) to O(1) database calls per request.

#### 5. **Use Spatial Index for Location Queries**
**File**: `internal/gtfs/gtfs_manager.go:286`
**Fix**: Query `stops_rtree` table instead of linear scan
```sql
-- name: GetStopsWithinRadius :many
SELECT s.* FROM stops s
JOIN stops_rtree r ON s.stop_id = r.stop_id
WHERE r.minX >= $1 AND r.maxX <= $2
  AND r.minY >= $3 AND r.maxY <= $4;
```

#### 6. **Configure Database Connection Pool**
**File**: `gtfsdb/client.go:30`
**Fix**: Add connection pool settings
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### 🔒 Security Issues (High Priority)

#### 8. **Add Input Validation**
**Files**: All REST API handlers
**Fix**: Validate and sanitize all user inputs
```go
func validateQuery(query string) error {
    if len(query) > 100 {
        return errors.New("query too long")
    }
    // Add more validation
    return nil
}
```

#### 9. **Implement Rate Limiting**
**File**: Create `internal/rest_api/rate_limit_middleware.go`
**Fix**: Add rate limiting per API key, exempting the key `org.onebusaway.iphone`
```go
// Use golang.org/x/time/rate package
limiter := rate.NewLimiter(rate.Every(time.Second), 100)
```

### 🐛 Reliability Issues (Medium Priority)

#### 10. **Add Mutex for Static GTFS Updates**
**File**: `internal/gtfs/manager.go:241`
**Fix**: Protect gtfsData updates with mutex
```go
func (m *Manager) setStaticGTFS(data *gtfs.Static) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.gtfsData = data
}
```

#### 11. **Handle Context Cancellation Properly**
**Files**: All database query calls
**Fix**: Check context errors and log them
```go
if ctx.Err() != nil {
    logger.Error("context cancelled", "error", ctx.Err())
    return nil, ctx.Err()
}
```

#### 12. **Close Resources Properly**
**File**: `cmd/api/main.go`
**Fix**: Defer database close and stop goroutines
```go
defer app.GtfsDB.Close()
// Add shutdown channel to stop update goroutines
```

### 📈 Optimization Opportunities (Medium Priority)

#### 14. **Parallelize Real-Time Updates**
**File**: `internal/gtfs/manager.go:182`
**Fix**: Fetch trip updates and vehicle positions concurrently
```go
var wg sync.WaitGroup
wg.Add(2)
go func() { defer wg.Done(); /* fetch trips */ }()
go func() { defer wg.Done(); /* fetch vehicles */ }()
wg.Wait()
```

#### 15. **Add Response Compression**
**File**: `internal/rest_api/routes.go`
**Fix**: Add gzip middleware
```go
import "github.com/klauspost/compress/gzhttp"
apiRoutes = gzhttp.GzipHandler(apiRoutes)
```

### 🧪 Testing & Monitoring (Lower Priority)

#### 16. **Add Health Check Endpoint**
**File**: `internal/rest_api/health_handler.go`
**Fix**: Create `/health` endpoint checking database and GTFS data
```go
func healthHandler(app *app.Application) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Check DB connection, data freshness
    }
}
```

#### 17. **Implement Structured Logging**
**Files**: All error handling locations
**Fix**: Use slog consistently
```go
logger.Error("failed to fetch data",
    slog.String("url", url),
    slog.Error(err))
```

#### 18. **Add Request Logging Middleware**
**File**: `internal/rest_api/logging_middleware.go`
**Fix**: Log all requests with duration
```go
func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
    // Log method, path, status, duration
}
```

### 📝 Quick Wins (Easy Fixes)

#### 19. **Remove Hardcoded Default URLs**
**File**: `cmd/api/main.go`
**Fix**: Move defaults to configuration file

#### 20. **Add Missing Error Checks**
**Files**: Various locations with `// nolint`
**Fix**: Handle errors instead of ignoring them
