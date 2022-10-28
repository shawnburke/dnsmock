# DNSMOCK

Mocking server and library.

## Server Usage

To use as an executable simply run the cmd package, with the following params:

* `--listen-addr`: Address to listen on, default is 0.0.0.0:0
* `--record`: Record DNS queries and responses, output them to stdout at exit
* `--record-file`: Record DNS queries and responses to specified file at exit.
* `--replay-file`: Replay the responses in the file (see below for details)
* `--downstreams`: Comma delimated list of downstreams or `localhost` (default) to load `/etc/resolv.conf`, or `none` to not have downstreams, e.g. anything not in replay file will fail to resolve.

```bash
go build -o dnsmock ./cmd
./dnsmock --listen-addr "0.0.0.0:50053" --record --downstreams "8.8.8.8,8.4.4.4"
```

Now test with

```bash
dig @0.0.0.0 -p 50053 google.com
```

When you exit this will output the queries and responses to stdout. Capture those to a file like "replay.yaml".

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

