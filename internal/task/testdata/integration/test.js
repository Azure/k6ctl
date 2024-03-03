import { Client } from "k6/net/grpc";

const client = new Client();
client.load(["/scripts"], "svc.proto");

export const options = {
  insecureSkipTLSVerify: true,
  // A number specifying the number of VUs to run concurrently.
  // vus: 1200,
  // A string specifying the total duration of the test run.
  // duration: "30s",

  stages: [
    { duration: '10s', target: 100 },
    { duration: '10s', target: 750 },
    { duration: '300s', target: 750 },
    { duration: '10s', target: 30 },
  ],
};

const token =
    "";

const saToken =
    "";

let connected = false;

export default function () {
}