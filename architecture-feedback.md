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

### 📈 Optimization Opportunities (Medium Priority)

#### 15. **Add Response Compression**
**File**: `internal/rest_api/routes.go`
**Fix**: Add gzip middleware
```go
import "github.com/klauspost/compress/gzhttp"
apiRoutes = gzhttp.GzipHandler(apiRoutes)
```

### 🧪 Testing & Monitoring (Lower Priority)

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

#### 20. **Add Missing Error Checks**
**Files**: Various locations with `// nolint`
**Fix**: Handle errors instead of ignoring them
