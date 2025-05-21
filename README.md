# sps30-exporter
sps30 sensor prometheus exporter



## Usage
### Docker-Compose

```yaml
  sps30:
    container_name: sps30
    image: bbcbear/sps30-exporter:1.0.0
    restart: unless-stopped
    ports:
      - '2112:2112'
    devices:
      - "/dev/i2c-1:/dev/i2c-1"
    environment:
      METRICS_ADDR: ":2112"
      POLL_INTERVAL: "2s"
    privileged: true
```
### Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'sps30'
    scrape_interval: 5s
    scrape_timeout: 3s
    static_configs:
      - targets: ['sps30:2112']
```
## Datasheet

- [SPS30 Particulate Matter Sensor Datasheet (Sensirion)](https://www.sensirion.com/file/datasheet/sps30-datasheet.pdf)
