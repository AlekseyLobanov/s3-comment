# s3-comment
I had few reasonris to write this project:
1. Something like guest book, because every developer should create it.
2. Is it possible to increase performance of Python application just with replacing it with Go?
And if yes, how painful it will be to write and how fast it may be.
3. Can we choose some exotic components in
our stack (S3-compatible server as comments storage) and achieve reasonable performance?
4. And also I need some golang experience

**Disclaimer:** it is not production-ready solution, it's mostly MVP.

## Features
Or What has been implemented for MVP.

- Basic API is compatibe with [isso](https://github.com/posativ/isso), including Markdown rendering
- Multiprocessing out of the box
- Multiple layers of caching (simple memory and redis) for effective and reliable caching with smart update.
Only necessary requests will be processed by S3.
- Prometheus metrics at `/metrics` endpoint with information for cache layers and API endpoints

## How to use
TBD

## Benchmarks
TBD