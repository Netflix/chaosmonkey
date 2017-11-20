A tracker is used to record termination events in some sort of external system.
Inside Netflix, we use trackers to record terminations to
[Atlas](https://github.com/netflix/atlas/wiki) (our metrics system) and to
Chronos, our event tracking system<sup>1</sup>.

If you wish to record terminations with some external system, you need to:

1. Give your tracker a name (e.g., "syslog")
1. Code up a type in Go that implements the [Tracker](https://godoc.org/github.com/Netflix/chaosmonkey/#Tracker) interface.
1. Modify [github.com/netflix/chaosmonkey/tracker/getTracker](https://github.com/Netflix/chaosmonkey/blob/master/tracker/tracker.go)
   so that it recognizes your tracker.
1. Edit your [config file](Configuration File Format) to specify your tracker.

---

<sup>1</sup>Unfortunately, we are unable to release either of these trackers as
open source. Our Atlas tracker communicates with a version of
[Prana](https://github.com/Netflix/Prana) that has not been released as open
source, and Chronos has also not been released as open source.
