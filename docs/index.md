---
layout: home

hero:
  name: "GOE Framework"
  text: "Modern Go Application Framework"
  tagline: Built on Uber's Fx & GoFiber for developer productivity and scalability
  image:
    src: /logo.svg
    alt: GOE Framework Logo
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: View Examples
      link: /examples/basic-app
    - theme: alt
      text: GitHub
      link: https://github.com/oeasenet/goe

features:
  - icon: 🔌
    title: Powerful Dependency Injection
    details: Built on Uber's Fx for type-safe management of application components and lifecycle with automatic dependency resolution.
  - icon: 🌐
    title: High-Performance HTTP Server
    details: Integrated with GoFiber v3 featuring automatic request logging, middleware support, and lightning-fast routing.
  - icon: 📝
    title: Structured & Flexible Logging
    details: Utilizes Uber's Zap logger with developer-friendly console output and production-ready JSON formatting.
  - icon: ⚙️
    title: Environment-Aware Configuration
    details: Load configuration from environment variables and .env files with type-safe accessors and hot-reloading support.
  - icon: 💾
    title: Versatile Cache Support
    details: Unified caching interface supporting multiple drivers like Memory, Redis, and SQLite for optimal performance.
  - icon: 🧩
    title: Extensible Module System
    details: Organize your application into logical modules with managed lifecycles (OnStart, OnStop) and clean separation of concerns.
  - icon: 🛡️
    title: Contract-Driven Design
    details: Core components are defined by interfaces, promoting loose coupling, testability, and maintainability.
  - icon: 🗄️
    title: GORM Database Integration
    details: Seamless integration with GORM for database operations, supporting MySQL, PostgreSQL, SQLite, and SQL Server.
  - icon: 📊
    title: Built-in Observability
    details: OpenTelemetry integration for metrics, tracing, and monitoring with Prometheus support out of the box.
---

