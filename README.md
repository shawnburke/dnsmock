# DNSMOCK
Capture and replay DNS results with a simple YAML configuration

![Go Test](https://github.com/shawnburke/dnsmock/actions/workflows/go.yml/badge.svg)  

[![Go Report Card](https://goreportcard.com/badge/github.com/shawnburke/dnsmock)](https://goreportcard.com/report/github.com/shawnburke/dnsmock)

[![codecov](https://codecov.io/gh/shawnburke/dnsmock/branch/main/graph/badge.svg)](https://codecov.io/gh/shawnburke/dnsmock)


DnsMock is a Go library that will return mocked DNS responses for any record type, in an easy to author YAML format.  Think of it as `/etc/hosts` on steroids, with added support for:

* Record resolution for all record types (not just A)
* Capturing and replaying DNS responses
* Wildcards and templating of query responses, for example capturing and responding to all queries for `*.foo.com`

It also can be run as a server that will record DNS results and return them as a YAML file, as well as replaying those results, including as a [Docker image](https://hub.docker.com/repository/docker/shawnb576/dnsmock).

Let's look at a simple example:

```yaml
# replay.yaml
rules:
 - name: google.com.
   records:
    A:
    - "google.com. 300 IN A 4.3.2.1"
```

Now run the server to return those results:

```bash
$./dnsmock -replay-file replay.yaml -port 9053 &
$ dig @localhost -p 9053 google.com

; <<>> DiG 9.16.1-Ubuntu <<>> @localhost -p 9053 google.com
; (2 servers found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 5540
;; flags: qr rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
;; WARNING: recursion requested but not available

;; QUESTION SECTION:
;google.com.                    IN      A

;; ANSWER SECTION:
google.com.             300     IN      A       4.3.2.1

```

This can also be run as a Docker container for easy portability.

## Server Usage Detail

To use as an executable simply run the cmd package, with the following params:

* `--port`: Port to listen on, default is 53
* `--record`: Record DNS queries and responses, output them to stdout at exit
* `--record-file`: Record DNS queries and responses to specified file at exit.
* `--replay-file`: Replay the responses in the file (see below for details)
* `--downstreams`: Comma delimated list of downstreams or `localhost` (default) to load `/etc/resolv.conf`, or `none` to not have downstreams, e.g. anything not in replay file will fail to resolve.

```bash
go build -o dnsmock ./cmd
./dnsmock --port 53" --record --downstreams "8.8.8.8,8.4.4.4"
```

Now test with

```bash
dig @0.0.0.0 -p 50053 google.com
```

When you exit this will output the queries and responses to stdout. Capture those to a file like "replay.yaml".

### Docker 

Note this is also available as a Docker image:

```bash
docker run -p "50053:53/udp" -d shawnb575/dnsmock:latest --record >replay.yml
dig @0.0.0.0 -p 50053 google.com
...
docker run -p "50053:53/udp"  -v "./replay.yml:/replay.yml" -d shawnb575/dnsmock:latest --replay-file /replay.yml
```




## Test Library Usage

You can also use just the library for standing up mock servers.

There is a working example usage in `TestMock` in the incredibly well named `e2e_example_mock_test.go` file.

To do this in your project, you can do something like:

```go
import (
    "github.com/shawnburke/dnsmock"
    "github.com/shawnburke/dnsmock/resolver"

    "go.uber.org/zap"
)

func runProxy() {
    logger := zap.NewNop()

    // create a resolver that looks up responses
    // from the replay file
    r := resolver.NewReplayFromFile("replay.yaml", logger)

    // Now create a new mocks server that serves those results
    p := dnsmock.New(":0", resolver, logger)
    err := p.Start()
    fmt.Println("Running at:", p.Addr())

    // Query it!
    client := &dns.Client{Network:"udp"}
    msg := &dns.Msg {
        Question: []dns.Question{
			{
				Name:   "google.com.",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
    }
    response, err := client.Exchange(msg, p.Addr())
    fmt.Println(response)
    p.Stop()
}

```

## Replay File Format

The replay file is simple, and allows wildcards. Entries are processed in order, first match wins.

Below, we set up rules for `internet.com` (exact match), `*.awstest.com`, and a fallthrough rule that returns `42` for all `SRV` requests.

```yaml
  rules:
    - name: "internet.com."
      records:
        A:
            - "internet.com.\t197\tIN\tA\t172.64.155.149"
            - "internet.com.\t197\tIN\tA\t104.21.31.107"
    - name: "*.awstest.com."
      records:
        A:
          - "{{Name}}\t300\tIN\tA\t1.2.3.4"
    - name: "*" 
      records:
        SRV:
        - "{{Name}}\t60\tIN\tSRV\t0 100 42 {{Name}}"

```

To get these values either record or copy them from `dig` output.

