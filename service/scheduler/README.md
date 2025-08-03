# Krypton Task Scheduler

## Schedule a task every so often
```Every()``` schedules a new periodic task with an interval. Interval can be an int, time.Duration or a string that parses with time.ParseDuration(). 

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

```
# Using time interval (int) format
_, _ = s.ForDevice({deviceID}, {tenantID}).Every(1).Second().Schedule(task)
_, _ = s.ForDevice({deviceID}, {tenantID}).Every(1).Minute().Schedule(task)
_, _ = s.ForDevice({deviceID}, {tenantID}).Every(1).Hour().Schedule(task)
_, _ = s.ForDevice({deviceID}, {tenantID}).Every(1).Day().Schedule(task)
_, _ = s.ForDevice({deviceID}, {tenantID}).Every(1).Days().Schedule(task)

# Using time.Duration format
_, _ = s.ForDevice({deviceID}, {tenantID}).Every(1 * time.Second).Schedule(task)

# Using time strings parseable by time.ParseDuration()
_, _ = s.ForDevice({deviceID}, {tenantID}).Every("1s").Schedule(task)
_, _ = s.ForDevice({deviceID}, {tenantID}).Every("1h").Schedule(task)
_, _ = s.ForDevice({deviceID}, {tenantID}).Every("24h").Schedule(task)
```

## Schedule a task at midday
```
_, _ = s.ForDevice({deviceID}, {tenantID}).Every(1).Day().Midday().Schedule(task)
```