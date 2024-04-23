import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
  vus: 1, // 1 virtual user
  duration: '15s', // for 10 seconds
};

export default function () {
  http.get('http://localhost:8080/hystrix');

  // Wait for 1 second before sending the next request
  sleep(1);
}
