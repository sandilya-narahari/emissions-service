# Emissions Cache Service

The **Emissions Cache Service** is a lightweight API gateway built to improve response times when retrieving emissions measurement data. It sits between a slow internal emissions API and customers, caching responses to ensure that key properties (with optional priority settings) are served in under 50ms. This service is written in Go and uses an in‑memory cache by default. The architecture is designed to be easily extended with additional observability and distributed caching if needed.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [API Endpoints](#api-endpoints)
- [Observability & Error Handling](#observability--error-handling)
- [Caching Strategy & Scalability](#caching-strategy--scalability)
- [Setup](#setup)
- [Testing](#testing)
- [Future Improvements](#future-improvements)

## Overview

The service accepts requests for emissions data by property (e.g., website domains or inventory IDs). It leverages a caching layer to:

- **Minimise latency:** Serve cached results in milliseconds.
- **Reduce load:** Fewer calls are made to the slow internal API.
- **Support prioritisation:** Customers can mark certain requests as priority, which will be cached indefinitely (no expiration) to ensure high availability.

### Key Steps in the Service Flow

1. **Accepts a request** at the `/v1/emissions/measure` endpoint.
2. **Checks the in‑memory cache** for an existing response using a composite key generated from country, channel, impressions, and inventory ID.
3. **Calls the internal Scope3 API** (if necessary) to fetch uncached emissions data.
4. **Stores responses** in the cache (using a longer TTL for regular requests and no expiration for priority requests).
5. **Returns the aggregated response** back to the client.

## Architecture

- **Config Loader:** Reads configuration from `config.yaml` with environment variable expansion.
- **HTTP Server:** Uses Gorilla Mux with middleware for request IDs and recovery from panics.
- **Handlers:** Route and process incoming HTTP requests.
- **Service Layer:** Contains business logic, including caching and API fallback logic.
- **Cache Repository:** Uses [go‑cache](https://github.com/patrickmn/go-cache) for in‑memory caching. The cache layer is abstracted via an interface so that you can easily swap it for a distributed cache (e.g., Redis) if scaling is needed.
- **Scope3 Client:** Manages calls to an external Scope3 API, with flexible configuration for timeouts and user agents.
- **Error Handling & Observability:** The service uses custom error types and structured logging (via standard logging with context propagation) to improve debugging and traceability. This design allows easy integration with observability tools like Sentry.
- **Tests:** Comprehensive unit tests are provided for each component.

## API Endpoints

### Health Check

**Endpoint:** `GET /v1/health`

**Response Example:**

```json
{
  "status": "healthy"
}
```

### Emissions Measurement

**Endpoint:** `POST /v1/emissions/measure`

**Request Payload Example:**

```json
{
  "rows": [
    {
      "country": "US",
      "channel": "online",
      "impressions": 1000,
      "inventoryId": "nytimes.com",
      "utcDatetime": "2025-01-01T12:00:00Z",
      "isPriority": false
    }
  ]
}
```

> **Note:**  
> - The `isPriority` flag indicates whether the cache entry should be permanent (no expiration).  
> - The composite cache key is generated using `country-channel-impressions-inventoryId`.

**Response Payload Example:**

```json
{
  "requestId": "api-req-001",
  "totalEmissions": 100.0,
  "rows": [
    {
      "propertyId": 1,
      "propertyName": "NyTimes Property",
      "totalEmissions": 100.0,
      "cached": false
    }
  ]
}
```

## Observability & Error Handling

- **Error Categorisation:** The service distinguishes between internal, validation, and external errors.
- **Structured Logging:** Each request is assigned a unique Request ID, and key events are logged with context to aid debugging.
- **Future Integration:** The error handling framework is designed to facilitate integration with external observability tools (e.g., Sentry, ELK stack).

## Caching Strategy & Scalability

- **In‑Memory Cache:** By default, the service uses an in‑memory cache with configurable TTL and cleanup intervals. The current implementation does not include eviction policies like LRU or capacity constraints.
- **Priority Caching:** Requests flagged as priority are cached without expiration. A planned enhancement will include background refresh of high-priority items to ensure data freshness.
- **Scalability Considerations:** The caching layer is abstracted via an interface so that you can easily replace it with a distributed caching solution like Redis when scaling horizontally.

## Setup

### Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [Docker](https://www.docker.com/get-started)

### Configuration

The service configuration is defined in `config.yaml`. Example configuration:

```yaml
scope3:
  api_url: "https://api.scope3.com/v2"
  token: "${SCOPE3_API_TOKEN}"
server:
  host: "0.0.0.0"
  port: 8080
cache:
  default_ttl: "24h"
  cleanup_interval: "1h"
```

### Running Locally

1. **Clone the repository:**

   ```bash
   git clone https://github.com/sandilya-narahari/emissions-service
   cd emissions-service
   ```

2. **Set environment variables:**

   ```bash
   export SCOPE3_API_TOKEN=your_scope3_api_token_here
   ```

3. **Build and start the container:**  
   From the root of the repository, run:
   
   ```bash
   docker-compose up --build
   ```

   This command will build the Docker image and start the container. Once running, the service will be accessible on port `8080`.

4. **To stop and remove the running containers, execute:**

   ```bash
    docker-compose down
    ```


## Testing

To run the unit tests:

```bash
go test ./... -v
```


## Future Improvements

### Observability & Monitoring
- **Enhanced Structured Logging & Contextual Error Reporting:**  
  Implement a comprehensive logging system that not only uses correlation IDs but also enriches error logs with contextual details (such as request IDs, input parameters, and operation metadata). This enables easier debugging and more actionable insights when errors occur.
- **Metrics Collection:**  
  Add Prometheus metrics to monitor cache hit/miss rates, API latencies, request volumes, and error frequencies.
- **Health Check Enhancements:**  
  Extend health checks to include detailed system status, uptime tracking, cache statistics, and the health of downstream dependencies.

### Resilience & Performance
- **Circuit Breaker Pattern:**  
  Implement circuit breakers in the Scope3 API client to handle external service failures gracefully and to prevent cascading issues.
- **Rate Limiting:**  
  Add configurable rate limiting to protect both the service and downstream systems from overload.
- **Cache Optimisation:**  
  - **Cache Coalescing:**  
    Utilise request deduplication (for example, using Go’s [`golang.org/x/sync/singleflight`](https://pkg.go.dev/golang.org/x/sync/singleflight)) to ensure that concurrent cache misses for the same key result in only one external API call. The returned result can then be shared among all waiting requests.
  - Implement cache warmup and preloading for frequently accessed keys.
  - Track cache statistics (hit rates, miss rates, eviction counts) to inform tuning and optimisation.
  - Support cache item prioritisation based on access patterns, ensuring that high-priority requests receive longer-lasting cache entries.

### Testing
- **Integration Tests:**  
  Develop a comprehensive integration test suite that validates the interactions between components, including cache behavior and external API calls.
- **Load Testing:**  
  Implement performance benchmarks and stress testing (using tools such as k6 or Locust) to ensure that the service meets its latency targets under high concurrency.
- **Chaos Testing:**  
  Introduce failure injection tests to verify the resilience of the service under adverse conditions.

### Architecture & Scalability
- **Distributed Caching:**  
  Replace or supplement the in-memory cache with a Redis-based caching solution to support horizontal scaling and multi-instance deployments.
