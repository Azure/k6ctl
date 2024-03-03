import { URL } from 'https://jslib.k6.io/url/1.0.0/index.js';
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  insecureSkipTLSVerify: true,

  stages: [
    { duration: "10s", target: 100 },
    { duration: "10s", target: 500 },
    { duration: "20s", target: 500 },
  ],
};

const message = __ENV.MESSAGE;

export default function () {
  const url = new URL('https://k6.io');
  url.searchParams.append('message', message);

  console.log(`requesting ${url.toString()}`);
  http.get(url.toString());

  sleep(1);
}

export function teardown() {
  http.post("http://localhost:4191/shutdown");
}