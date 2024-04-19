import http from 'k6/http';
import { sleep, check } from 'k6';

export let options = {
  duration: '10s',  // Run the test for 10 seconds
};

let successfulRequests = 0;
let failedRequests = 0;

export default function () {
  // Send 100 requests to the server
  for (let i = 0; i < 100; i++) {
    let response = http.get('http://localhost:8080/hystrix');
    let checkResult = check(response, {
      'status is 200': (r) => r.status === 200,
    });
    if (checkResult) {
      successfulRequests++;
    } else {
      failedRequests++;
    }
  }
  
  // Wait for 10 seconds before sending the next batch of requests
  sleep(10);
}

export function handleSummary(data) {
  console.log(`Successful requests: ${successfulRequests}`);
  console.log(`Failed requests: ${failedRequests}`);
}