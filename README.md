# High-Performance Web Crawler in Go

| CI Status                                                                                                                | Go Version                                                                                               | License                                                                                                   |
| :----------------------------------------------------------------------------------------------------------------------: | :------------------------------------------------------------------------------------------------------: | :---------------------------------------------------------------------------------------------------------: |
| [![CI](https://github.com/NesterovYehor/Crawler/actions/workflows/ci.yml/badge.svg)](https://github.com/NesterovYehor/Crawler/actions/workflows/ci.yml) | ![Go 1.23+](https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat-square&logo=go) | ![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square) |

A concurrent, high-performance web crawler built in Go, designed for scalability and resilience. This project features a sophisticated architecture using Redis for caching and queuing, Cassandra for metadata storage, and Prometheus for real-time performance monitoring.

---

## 📖 Overview

This crawler is built on a distributed, microservice-inspired architecture designed for high throughput. The core of the system is a **concurrent worker pool** that processes tasks from a **multi-priority queue** in Redis. To ensure politeness and avoid overwhelming servers, a **token-bucket rate limiter** (implemented as a Redis Lua script) is used.

The storage layer is designed for scale, using **Cassandra** for metadata, **S3** for raw content, and a **Bloom Filter in Redis** to prevent re-crawling duplicate URLs. The entire system is instrumented with **Prometheus** metrics, which can be visualized in **Grafana** to provide real-time insight into the crawler's performance.

---

<details>
<summary>🏛️ Click to Expand: Detailed Architecture & Design Decisions</summary>

<br>

The core logic is separated into two main processes: fetching and storing, each handled by a dedicated type of worker.

<!-- 
======================================================================
 VVVV    PASTE YOUR TWO HIGH-LEVEL DIAGRAMS IN THE TABLE BELOW     VVVV
====================================================================== 
-->
<table>
  <tr>
    <td>
      <p align="center"><b>Fetch Process Flow</b></p>
      <img width="1631" height="1117" alt="High_level" src="https://github.com/user-attachments/assets/f7451171-857c-45d6-b6f8-6926986b0ee7" />
    </td>
    <td>
      <p align="center"><b>Store Process Flow</b></p>
      <img width="1649" height="694" alt="Store" src="https://github.com/user-attachments/assets/d34a797b-8315-4d41-b799-dad6f647a0ac" />
    </td>
  </tr>
</table>

The key design decisions were made to ensure scalability, resilience, and high performance:

- **High Concurrency with a Worker Pool:** The core of the crawler is a pool of goroutine workers. This design maximizes CPU utilization by processing multiple pages in parallel and allows the application to be scaled horizontally by simply increasing the number of worker instances.

- **Distributed Caching & Queuing with Redis:** Redis was chosen for its high-performance, in-memory data structures.
    - **Task Queues:** Redis Streams are used to implement a durable, multi-priority task queue. This allows for reliable task distribution and ensures that high-priority work (like processing sitemaps) is handled first.
    - **Duplicate Prevention:** A Bloom Filter, a memory-efficient probabilistic data structure, is used to keep track of all visited URLs. This dramatically reduces redundant work and saves storage resources.

- **Scalable Storage Layer:** The storage system is designed to handle a massive volume of write operations.
    - **Metadata (Cassandra):** A Cassandra cluster was chosen for storing page metadata. Its masterless architecture and high write throughput are ideal for a write-heavy application like a web crawler.
    - **Content (S3):** The raw HTML content of each page is saved to an S3-compatible blob store, which provides cheap, durable, and highly available storage for large objects.

- **Polite Crawling:** To ensure the crawler is a good citizen of the web, a sophisticated politeness manager was implemented. It respects `robots.txt` rules (cached in Redis) and uses a token-bucket rate limiting algorithm. This algorithm is implemented in a **Redis Lua script** to guarantee atomic check-and-decrement operations, preventing race conditions in a highly concurrent environment.

- **Real-Time Monitoring:** The application exposes detailed performance metrics via a `/metrics` endpoint. This allows for real-time monitoring with Prometheus and visualization in Grafana, providing crucial insights into the crawler's health and performance.

### Queue Picking Flow

The worker pool uses an intelligent, priority-based scheduling system to ensure that high-priority tasks are handled first, while also maximizing worker utilization. Each worker is initialized with a primary queue to listen to, but the pool can dynamically assign tasks from other queues.

The logic is as follows: When a worker requests a new task, the worker pool first attempts to pull a task from the worker's assigned primary queue. If that queue is empty, the pool will scan all queues in a fixed fallback order (`High` > `Medium` > `Store` > `Retry`) and assign the first available task it finds. This "work-stealing" approach ensures that no worker sits idle as long as there is work to be done anywhere in the system.

<!-- 
======================================================================
 VVVV    PASTE YOUR "QUEUE PICKING FLOW" DIAGRAM HERE     VVVV
====================================================================== 
-->
<img width="1581" height="734" alt="queue" src="https://github.com/user-attachments/assets/5f97c03a-2ebe-4c28-9291-0cc42d32cef3" />

### Politeness Flow

To ensure the crawler is a good citizen of the web, it uses an event-driven politeness flow with a clean separation of concerns. The worker, not the politeness manager, is responsible for initiating the fetch of a new `robots.txt` file.

The flow is as follows: A worker asks the politeness manager for permission to crawl a URL. The manager executes a single, atomic Lua script on Redis that checks the rate-limit tokens and returns the cached `robots.txt` data if available. If the Redis key for the domain does not exist, the manager returns a `redis.Nil` error. The worker interprets this error as a signal to create a new, high-priority `fetch_rules` task and sends it back to the queue. This decoupled, event-driven design prevents the politeness manager from making network calls itself and keeps the system robust.

<!-- 
======================================================================
 VVVV    PASTE YOUR "POLITENESS FLOW" DIAGRAM HERE     VVVV
====================================================================== 
-->
<img width="1254" height="727" alt="Снимок экрана 2025-08-02 в 14 32 25" src="https://github.com/user-attachments/assets/47ae0791-d1a3-4d6a-b426-1068c488128b" />

</details>

---

## 📊 Performance Highlights

The crawler is capable of processing thousands of pages in minutes on a single machine. The primary bottleneck is intentionally the network I/O and politeness delays, not the application's processing logic.

<!-- 
======================================================================
 VVVV    PASTE A SCREENSHOT OF YOUR "CRAWL RATE" GRAFANA PANEL HERE     VVVV
====================================================================== 
-->
<img width="2067" height="485" alt="Снимок экрана 2025-08-05 в 15 32 07" src="https://github.com/user-attachments/assets/c05d667f-f8e0-4bf6-96a2-76ffba2598ac" />

<!-- 
======================================================================
 VVVV    PASTE A SCREENSHOT OF YOUR "CRAWL DURATION" PANEL HERE     VVVV
====================================================================== 
-->

<img width="2204" height="913" alt="Снимок экрана 2025-08-05 в 13 10 13" src="https://github.com/user-attachments/assets/3f1f5676-a8ba-47c2-8f55-983227b2c454" />


> **Note on Performance:** The metrics shown above were generated by running the crawler on a single machine against a diverse set of websites. Your actual results may vary depending on your hardware, network conditions, and the websites you choose to crawl.

---

## 🚀 Getting Started

This project is fully containerized using Docker and Docker Compose for easy setup.

### Prerequisites

- Docker and Docker Compose

### Running the Crawler

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/NesterovYehor/Crawler.git](https://github.com/NesterovYehor/Crawler.git)
    cd Crawler
    ```

2.  **Build and run the services:**
    ```bash
    docker compose up --build
    ```
    This will start the crawler, Redis, Cassandra, Prometheus, and Grafana containers.

3.  **View the metrics:**
    - The crawler's metrics are exposed at `http://localhost:2112/metrics`.
    - The Prometheus UI is available at `http://localhost:9090`.
    - The Grafana dashboard is available at `http://localhost:3000` (login with `admin`/`admin`).

---

## 🛠️ Configuration

The application is configured via the `config.yaml` file. Key options include worker pool size, concurrency limits, and database connection details. Environment variables are used within the `docker-compose.yml` file to correctly wire the services together.
