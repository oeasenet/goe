---
layout: home

hero:
    name: "GOE Framework"
    text: "Build Production Go Apps in Minutes"
    tagline: Dependency injection, HTTP, databases, caching, jobs, and distributed locking — wired together and ready to go. Built on Uber Fx and GoFiber v3.
    image:
        src: /goe_gopher_logo.png
        alt: GOE Framework Logo
    actions:
        -   theme: brand
            text: Get Started
            link: https://deepwiki.com/oeasenet/goe
        -   theme: alt
            text: View on GitHub
            link: https://github.com/oeasenet/goe
        -   theme: alt
            text: Browse Examples
            link: https://github.com/oeasenet/goe/tree/v2/examples

features:
    -   icon: 🔌
        title: Dependency Injection
        details: Built on Uber Fx for type-safe component wiring with automatic resolution and lifecycle management. No globals, no init() magic.
    -   icon: 🌐
        title: High-Performance HTTP
        details: GoFiber v3 with automatic request logging, middleware support, and one of the fastest routing engines in the Go ecosystem.
    -   icon: 📝
        title: Structured Logging
        details: Uber Zap under the hood — colored console output for development, structured JSON for production. Injected everywhere automatically.
    -   icon: ⚙️
        title: Environment-Aware Config
        details: Layered .env files with environment overrides, type-safe accessors, and validation. No more missing config surprises at runtime.
    -   icon: 🗄️
        title: SQL & MongoDB
        details: GORM integration for SQL databases (MySQL, PostgreSQL, SQLite, SQL Server) and native MongoDB driver with pooling, transactions, and migrations.
    -   icon: 💾
        title: Caching with 10+ Backends
        details: Unified caching interface supporting Redis, Memory, Memcache, Badger, SQLite3, PostgreSQL, MySQL, MongoDB, DynamoDB, and S3.
    -   icon: 📋
        title: Background Jobs
        details: Redis-backed job processing with human-friendly scheduling, delayed execution, retries with exponential backoff, and dead letter queues.
    -   icon: 🔐
        title: Distributed Locking
        details: Redis-based mutex with Redlock algorithm support for single instance, Sentinel, and Cluster deployments.
    -   icon: 🧩
        title: Module System
        details: Organize your app into modules with managed lifecycles (OnStart, OnStop), config validation, and clean separation of concerns.
    -   icon: 🛡️
        title: Contract-Driven Design
        details: Every core component is defined by an interface. Swap implementations, mock in tests, and keep coupling low across your codebase.
    -   icon: ⚡
        title: Concurrency Safe
        details: Singleflight cache protection, atomic operations, and double-check locking throughout. Race-free by default, verified with -race in CI.
    -   icon: 🩺
        title: Observability Ready
        details: Built-in health checks, OpenTelemetry tracing integration, and structured metrics. Production visibility from day one.
---
