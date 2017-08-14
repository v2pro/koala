# koala

人生如戏，全凭演技

# Parameters

* KOALA_MODE: REPLAYING/RECORDING

# Recording

![recording](https://docs.google.com/drawings/d/1IRmc6LH4tLq9l8ELF2XaGouzqr51Hb-0n2QN25zpiEg/pub?w=669&h=471)

* intercept tcp send/recv
* associate send/recv data to same thread id as "session"
* request => response => request, so we can know when a "talk" (with request/response pair) is complete
* use udp 127.127.127.127:127 to inform recorder with helper information.
* for "system under test" using thread multiplexing (one thread doing more than one thing), 
map real thread id to virtual thread id by helper information.

# Replaying

![replaying](https://docs.google.com/drawings/d/1uTW-4Hedimy4mLGTQtCG5lDLrmYfWXMZm6PfuabRdYY/pub?w=960&h=720)

replaying builds on same mechanism, but much more complex

* "system under test" is a process, "replayer inbound server" and "replayer outbound server" lives in same process. 
They are two tcp servers started by the .so loaded via LD_PRELOAD.
* session to replay is injected into the process via "replayer inbound server" tcp connection
* "replayer inbound server" call the "system under test" via tcp connection, store the "session id <=> inbound socket" mapping.
* "system under test" call external dependencies, which is intercepted to "replayer outbound server", store the "inbound socket <=> outbound socket" mapping
* "replayer outbound server" use its own socket to lookup the mapping, to find which session to replay

# Real World Scenarios

* Long Connection: reusing connection sequentially is not a issue, just update the mapping
* Multiplexing: one thread handing multiple business processes at the same time, need "helper information"
* One Way Communication: request without response, need "helper information" to cut two requests out
* Greeting: protocol like mysql send greeting before request. use the ip:port or just guess, to decide if greeting is needed.

