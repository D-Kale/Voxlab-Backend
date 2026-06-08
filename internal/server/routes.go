package server

// routes.go - Declarative route definitions
//
// All API routes are defined here in one place.
// Pattern: Next.js declarative + Laravel organized by function.
//
// Current routes:
//   GET    /api/v1/health              → HealthCheck
//   POST   /api/v1/auth/login          → Login
//   GET    /api/v1/tracks              → GetTracks
//
// To add a new route:
//   1. Create handler in the appropriate domain package
//   2. Add the route here

// Route documentation is kept in individual handler files
// with Swagger annotations where needed.
