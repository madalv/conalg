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

function chooseName(probOfConflict) {
  if (Math.random() * 100 < probOfConflict) {
    const randomIndex = Math.floor(Math.random() * keyPool.length);
    return keyPool[randomIndex];
  }
  return `command${Math.floor(Math.random() * 50)}`;
}

export const options = {
  stages: [
    { duration: '15s', target: 10 },  // Ramp up to 50 VUs over 30 seconds
    //{ duration: '30s', target: 100 }, // Ramp up to 100 VUs over next 30 seconds
    //{ duration: '30s', target: 60 }, // Ramp up to 150 VUs over next 30 seconds
    //{ duration: '30s', target: 200 }, // Ramp up to 200 VUs over next 30 seconds/
    // { duration: '30s', target: 100 }, // Ramp down to 100 VUs over next 30 seconds
    // { duration: '30s', target: 50 },  // Ramp down to 50 VUs over next 30 seconds
    // { duration: '30s', target: 0 },   // Ramp down to 0 VUs over next 30 seconds
  ],
};

export default function () {
  const randomIndex = Math.floor(Math.random() * urls.length);
  const url = urls[randomIndex];
  const payload = chooseName(70);

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