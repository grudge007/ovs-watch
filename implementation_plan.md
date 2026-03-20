# ovs-watch

A CLI tool that watches Open vSwitch in real time and prints changes as they happen.

---

## What it does

You run it on any node that has OVS running.
It polls OVS every N seconds, compares current state to previous state,
and prints only what changed — nothing if nothing changed.

```
$ ovs-watch

[14:23:01] [ADDED]   bridge   br-int
[14:23:45] [ADDED]   port     vxlan-c0a80102  on br-int
[14:24:10] [REMOVED] port     eth0            on br-ext
[14:25:00] [CHANGED] port     vxlan-c0a80102  state: up → down
[14:25:33] [ADDED]   bridge   br-ext
```

Nothing else. No noise. Just changes.

---

## Requirements

### R1 — Watch bridges
Detect when a bridge is added or removed.

```
[14:23:01] [ADDED]   bridge  br-int
[14:23:10] [REMOVED] bridge  br-ext
```

### R2 — Watch ports
Detect when a port is added or removed on any bridge.
Show which bridge it belongs to.

```
[14:23:45] [ADDED]   port  vxlan-c0a80102  bridge=br-int
[14:24:10] [REMOVED] port  eth0            bridge=br-ext
```

### R3 — Watch interface state
Detect when a port's link state changes (up/down).

```
[14:25:00] [CHANGED] port  vxlan-c0a80102  state: up → down
```

### R4 — Watch tunnel endpoints
For VXLAN/Geneve ports specifically, detect when remote_ip or VNI changes.

```
[14:26:00] [CHANGED] tunnel  vxlan-c0a80102  remote_ip: 10.0.0.1 → 10.0.0.2
```

### R5 — Configurable poll interval
Default 2 seconds. User can change it.

```
ovs-watch --interval 5
```

### R6 — Filter by bridge
Optionally watch only one bridge instead of all.

```
ovs-watch --bridge br-int
```

### R7 — Filter by event type
Optionally show only adds, only removes, or only changes.

```
ovs-watch --events added,removed
ovs-watch --events changed
```

### R8 — Output modes
Two output modes:

**pretty** (default) — human readable, colored, with timestamps
```
[14:23:01] [ADDED]   bridge  br-int
```

**json** — one JSON object per line, for piping to other tools
```json
{"time":"14:23:01","event":"added","type":"bridge","name":"br-int"}
{"time":"14:23:45","event":"added","type":"port","name":"vxlan-c0a80102","bridge":"br-int"}
```

```
ovs-watch --output json
```

### R9 — Graceful shutdown
Ctrl+C exits cleanly. No panic, no hanging goroutines.
Print a summary on exit:

```
^C
ovs-watch stopped. watched for 4m32s. 12 events captured.
```

### R10 — OVS not running
If OVS is not running when the tool starts, print a clear error and exit.
If OVS goes down while watching, print a warning and keep retrying — don't crash.

```
ERROR: OVS not reachable. Is ovs-vswitchd running?
```

```
[14:30:00] [WARN] OVS unreachable, retrying...
[14:30:02] [WARN] OVS unreachable, retrying...
[14:30:04] OVS reconnected, resuming watch
```

---

## What you are NOT building

- No TUI / dashboard (no bubbletea, no ncurses)
- No persistent storage or log files
- No remote watching over SSH
- No flow table watching (too complex for now)
- No metrics or Prometheus endpoint

Keep it simple. It is a focused single-purpose tool.

---

## CLI interface summary

```
ovs-watch [flags]

Flags:
  --interval int      poll interval in seconds (default 2)
  --bridge  string    watch only this bridge (default: all bridges)
  --events  string    comma separated: added,removed,changed (default: all)
  --output  string    output format: pretty or json (default: pretty)
  --help              show help
```

---

## Go concepts this will teach you

Work through the requirements in order.
Each one introduces something new.

| Requirement | Go concept you will learn |
|---|---|
| R1 — watch bridges | structs, methods, exec.Command, string parsing |
| R2 — watch ports | slices, maps, nested structs |
| R3 — watch state | comparing structs, detecting diffs |
| R4 — watch tunnels | embedding structs, extending existing types |
| R5 — poll interval | time.Ticker, goroutines basics |
| R6, R7 — filters | function arguments, filtering slices |
| R8 — output modes | interfaces — your first real interface |
| R9 — graceful shutdown | context, os.Signal, channel select |
| R10 — error handling | error wrapping, retry loops |

---

## Suggested file structure

Start with everything in one file. Split only when a file crosses ~150 lines.

```
ovs-watch/
├── main.go        ← CLI flags, setup, main loop
├── ovs.go         ← OVS polling, structs for bridge/port/interface state
├── diff.go        ← comparing old state vs new state, producing events
├── output.go      ← pretty printing and JSON output
└── go.mod
```

---

## Definition of done

You know it is working when you can:

1. Run `ovs-watch` on a node
2. In another terminal run `ovs-vsctl add-br test-br`
3. See `[ADDED] bridge test-br` printed immediately (within poll interval)
4. Run `ovs-vsctl del-br test-br`
5. See `[REMOVED] bridge test-br` printed
6. Press Ctrl+C and see the summary line

That is the full loop. If that works, the tool works.