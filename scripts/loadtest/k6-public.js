// k6 script for StreamPass public API (BL-032, optional).
// Requires k6: https://k6.io
//
//   k6 run scripts/loadtest/k6-public.js
//   k6 run -e BASE=https://212-43-156-33.nip.io scripts/loadtest/k6-public.js

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10,
  duration: '20s',
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
  },
};

const BASE = __ENV.BASE || 'https://212-43-156-33.nip.io';

export default function () {
  const paths = ['/health', '/api/v1/rules', '/api/v1/config', '/api/v1/regions'];
  const path = paths[Math.floor(Math.random() * paths.length)];
  const res = http.get(`${BASE}${path}`);
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
  sleep(0.2);
}
