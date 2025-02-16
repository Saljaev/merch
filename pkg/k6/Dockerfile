FROM grafana/k6:latest

COPY script.js /test.js

CMD ["run", "--summary-export=results.json", "/test.js"]
