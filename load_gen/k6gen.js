import http from 'k6/http';
import { sleep } from 'k6';

const urls = [
  "http://localhost:8081/propose",
  "http://localhost:8082/propose",
  "http://localhost:8083/propose",
  // Add more URLs if needed
];

const keyPool = [
  "key1",
  "key2",
  "key3",
  "key4",
  "key5",
];

function chooseName() {

  return `command${Math.floor(Math.random() * 50)}`;
}

export const options = {
  scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 300,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 300,
    },
  },
};

export default function () {
  const randomIndex = Math.floor(Math.random() * urls.length);
  const url = urls[randomIndex];
  const payload = chooseName();

  console.log(`Sending request to ${url} with payload ${payload}`);

  const data = {
    command: payload,
  };

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  http.post(url, JSON.stringify(data), params);
}