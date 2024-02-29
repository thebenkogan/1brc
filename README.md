### 1 Billion rows challenge in Go
My solution to [gunnarmorling/1brc: 1️⃣🐝🏎️ The One Billion Row Challenge](https://github.com/gunnarmorling/1brc)

Contains two binaries:
- `create_measurements`: Scrapes Wikipedia for average yearly temperatures of cities around the world, producing a dataset of measurements of the specified size.
- `calculate`: Computes the min/max/mean/count temperature of each city.

For a dataset of 1 billion measurements, `calculate` runs in under 20 seconds on my 16-core machine.

I use a rather simple approach utilizing as many CPU cores as possible to break the file into chunks and produce the city statistics for each chunk. I then merge the results together to obtain the final statistics. The line parser function was also optimized a bit, which I noticed from profiling.
