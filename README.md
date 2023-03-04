# spod
(S)parkling (P)rometheus (U)gly (D)iff

## Wut?

Inspired by a very real scenario which came up recently, I found myself wondering
"why can't I hold all of this `diff --suppress-common -y <(some-prometheus-exporter-endpoint | sort) <(ditto | sort)` output?"
and thought maybe I could do something to tease a few brain cells I have back from the ledge of apoptosis...

The gist is, it's a damned pain to diff two disparate Prometheus/OTtel metrics exporter endpoints and draw useful* conclusions.
Therefore, I try to do better for this very niche concern by:
- Taking two endpoints
- Collecting their metrics
- Displaying a diff, especially when the values by a certain input amount
