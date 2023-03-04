<<<<<<< Updated upstream
# spod
<<<<<<< Updated upstream
(S)parkling (P)rometheus (U)gly (D)iff

=======
(Sp)arkling (O)penTelemetry (metrics) Diff
=======
# spud
(S)parkling (P)rometheus (U)gly (D)iff

![example showing several lines of columnar output. The columns, left-to-right signify: Timeseries name (first exporter),
Timeseries value (first exporter), one of: <. >, | to signify if the series is found in the first, second, or both exporters.
The next two columns represent the timeseries name and value for the second exporter. The last column signifies absolute
difference between the two values, if applicable. One line is bolded to draw attention to the timeseries existing in both
exporters, but with different values.](example.png)

>>>>>>> Stashed changes
## Wut?

Inspired by a very real scenario which came up recently, I found myself wondering
"why can't I hold all of this `diff --suppress-common -y <(some-prometheus-exporter-endpoint | sort) <(ditto | sort)` output?"
and thought maybe I could do something to tease a few brain cells I have back from the ledge of apoptosis...

The gist is, it's a damned pain to diff two disparate Prometheus/OTtel metrics exporter endpoints and draw useful* conclusions.
Therefore, I try to do better for this very niche concern by:
- Taking two endpoints
- Collecting their metrics
- Displaying a diff, especially when the values by a certain input amount
<<<<<<< Updated upstream
=======
>>>>>>> Stashed changes
>>>>>>> Stashed changes
