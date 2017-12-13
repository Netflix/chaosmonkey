## Enabled group

Chaos Monkey will only consider server groups eligible for termination if they
are marked as enabled by Spinnaker.  The Spinnaker API exposes an *isDisabled*
boolean flag to indicate whether a group is disabled. Chaos Monkey filters on
this to ensure that it only terminates from active groups.

## Probability

For each app, Chaos Monkey divides the instances into instance groups (the groupings
depend on how the app is configured). Every weekday, for each instance group,
Chaos Monkey flips a weighted coin to decide whether to terminate an instance
from that group. If the coin comes up heads, Chaos Monkey schedules a termination at
a random time between 9AM and 3PM that day.

Under this behavior, the number of work days between terminations for an
instance group is a random variable that has a [geometric distribution][1].

The equation below describes the probability distribution for the time between
terminations. *X* is the random variable, *n* is the number of work days between
terminations, and *p* is the probability that the coin comes up heads.

    P(X=n) = (1-p)^(n-1) × p,   n>=1

Taking expectation over *X* gives the mean:

    E[X] = 1/p

Each app defines two parameters that governs how often Chaos Monkey should terminate
instances for that app:

 * mean time between terminations in work days (μ)
 * min time between terminations in work days  (ɛ)

Chaos Monkey uses μ to determine what *p* should be. If we ignore the effect of
ɛ and solve for *p*:

    μ = E[X] = 1/p
    p = 1/μ

As an example, for a given app, assume that μ=5. On each day, the probability of
a termination is 1/5.

Note that if ɛ>1, Chaos Monkey termination behavior is no longer
a geometric distribution:

    P(X=n) = (1-p)^(n-1) × p,  n>=ɛ


In particular, as ɛ grows larger, E[X]-μ gets larger. We don't apply a
correction for this, because the additional complexity in the math isn't worth
having E[X] exactly equal μ.

Also note that if μ=1, then p=1, which guarantees a termination each day.



[1]: https://en.wikipedia.org/wiki/Geometric_distribution
