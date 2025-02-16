import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { Trend } from 'k6/metrics';

const responseTimeTrend = new Trend('response_time_trend');

export let options = {
    scenarios: {
        constant_rps: {
            executor: 'constant-arrival-rate',
            rate: 1000,
            timeUnit: '1s',
            duration: '5m',
            preAllocatedVUs: 20,
            maxVUs: 1000,
        },
    },
    thresholds: {
        http_req_duration: [
            'p(99) < 1000',
            'avg < 500',
            'max < 2000'
        ],
        http_req_failed: ['rate < 0.01'],
        checks: ['rate >= 0.99']
    },
    summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(95)', 'p(99)'],
};

export default function () {
    let payload = JSON.stringify({
        username: randomString(30),
        password: randomString(30),
    });

    let res = http.post('http://backend:8080/api/auth', payload, {
        headers: {'Content-Type': 'application/json'},
    });

    responseTimeTrend.add(res.timings.duration);

    check(res, {
        'status is 200': (r) => r.status === 200,
    });
}