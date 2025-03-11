# **Crawler – High-Performance Web Crawler**

📌 **Efficient, Concurrent, and Scalable Web Scraper built with Go**

[![Go Version](https://img.shields.io/badge/Go-1.20%2B-blue)](https://golang.org/)  
[![License](https://img.shields.io/github/license/NesterovYehor/Crawler)](LICENSE)  

---

## **📖 Overview**
**Crawler** is a high-performance web scraping tool written in **Go**, leveraging **goroutines, channels, and mutexes** for efficient, concurrent crawling. It extracts **URLs, text, and metadata** from web pages while maintaining rate limits and error resilience.

---

## **🚀 Features**
✅ **Concurrent Crawling** – Uses **goroutines & channels** to fetch multiple pages at once.  
✅ **Efficient URL Parsing** – Extracts and normalizes links using [`golang.org/x/net/html`](https://pkg.go.dev/golang.org/x/net/html).  
✅ **Rate-Limiting & Throttling** – Avoids overwhelming target servers.  
✅ **Resilient Error Handling** – Retries failed requests intelligently.  

---

## **📦 Installation**
Make sure you have **Go 1.20+** installed. Then, clone the repository and build the project:

```sh
# Clone the repository
git clone https://github.com/NesterovYehor/Crawler.git
cd Crawler

# Build the project
go build -o crawler main.go
```

---

## **🛠 Usage**
Run the crawler with a starting URL:

```sh
./crawler -url=https://example.com -depth=2
```

### **Command-Line Options:**
| Flag            | Description                                   |
|----------------|-----------------------------------------------|
| `-url`        | The starting URL for crawling                 |
| `-depth`      | Maximum depth for recursive crawling          |
| `-workers`    | Number of concurrent workers                  |
| `-timeout`    | Timeout for HTTP requests (in seconds)        |


---

## **📜 License**
This project is licensed under the **MIT License** – see the [LICENSE](LICENSE) file for details.

---

## **📬 Contact**
👤 **Yehor Nesterov**  
📧 yehorn38@gmail.com  
📌 [GitHub](https://github.com/NesterovYehor)  

---

🚀 **Happy Crawling!**
