# Exploring net/http package in Go to build web servers. Is it enough, or do we need frameworks?

And how easy is it to use? Does it require a lot of boilerplate?  
Following [this video](https://www.youtube.com/watch?v=H7tbjKFSg58) to understand new features; if net/http ends up being as feature rich as Express, I call no need for frameworks. I have another repo where I'm following an in-depth Express framework tutorial and go over all important features.

[Jump down to it](#guess-the-number---tcpudpquic) Also doing some experimenting with TCP/UDP/QUIC and HTTP/2.0, 3.0 in Go, over in cmd/[tcp/udp/quic?]/main.go
Implementing a guess-the-number game with client-server. Notes (requirements and learning) at the bottom.

### Perspective from which this project is written

- I played more with Nodejs-based frameworks like Nextjs, Express, NestJS so that's where I'm coming from
- I do have a decent understanding of how http requests are handled and how we can use middleware and define a pipeline to process our requests and responses
- I'm comparing features I've seen in Express with what net/http can do and then seeing if frameworks like Gin, Echo, Fiber, etc. are worth it

### Features of net/http

- first of all, defining & starting a server on a specified port and routing requests based on simple paths; this is pretty direct
- path parameters also seem to be easy to implement; but this is only available starting with Go 1.22 (and current version is 1.23.2, so not that long ago); before this, some manual labor was necessary, or introducing a framework; so + for frameworks, but also + for Go projects which can upgrade to this version
  - careful with conflicting paths and precedence; this is a problem in Express; (in Express) if I have "/users/{id}" and "/users/items", the second one will never be reached because the first one will match first, as long as we don't impose other rules regarding what the id looks like; the solution would be to put the "/users/items" route first
  - in Go, we don't worry about that; net/http has a rule "most specific wins", so no matter the order the more specific route handler will be called
  - but if we have a situation like "/users/{id}" and "/{resource}/items"? there's the same level of specificity with 1 parameter; in this case, Go detects it and panics (equivalent of throwing an error); _cool thing about Go's panic mechanism is that deferred functions still execute (so there's an implied "finally" block)_; btw, panics are not exceptions/errors, Go is meant to use errors for that; panics are for unrecoverable situations
- method-based routing - like in Express, specifying for the same route to be handled differently based on the HTTP method; this is a must-have feature
  - prior to 1.22, we would need to manually check (so boilerplate); now we prepend the method to the route, like _router.handleFunc("GET /users", getUsers)_;
  - so again, + for frameworks prior to 1.22, but + for net/http after 1.22
  - if we have a "PUT /users" and "/users", the non-method-specific route will encompass the other methods (so again, specificity); so it's completely handled
  - **when specifying the method in the route, only have 1 space between the method and the path; otherwise, it won't work**
- (personal note) if for any reason someone couldn't handle the idea of writing "handleFunc" instead of ".use"/".get" / the other Express-like things, a bit of boilerplate/wrapping could surely fix that; but this should never happen
- host-based routing - can specify the host inside the route string; I personally haven't seen this being used, though I think it's useful when trying to handle subdomains, like when website.com wants to do blog.website.com instead of website.com/blog; so that's a +; I don't see 1.22 update specifying implementing this, so it might've been there before
- TLS - not sure if it didn't have this before, but it surely has it post-1.22; so we can easily serve our server over HTTPS, that's a +
- **middleware** - next most important thing after being able to define route handlers
  - definitely works by defining a function that takes a handler and returns a handler; inside, we can define our logic by returning a handlerFunc and implementing our logic inside its closure
  - issue: we'd need to nest functions inside functions inside functions...; definitely not scalable
  - solution is to chain Middleware using a stack; we push our middleware functions to this stack and the chain handles that Lisp-like sandwich of functions
  - another issue is that we need boilerplate to obtain the response status code; so if we want to log it for example just before sending, we need to wrap the response writer and add our own fields, which we then need to start using inside our route handling functions; this chains us to our own implementation of the response writer wrapper, which requires big refactors later down the line if we want to change things
  - that's a - for net/http and + for frameworks due to these 2 issues of middleware chaining and obtaining status code...**is what I would say but I did not find an easy way of obtaining status code on Chi for example**; Chi has a simple way of chaining middleware; Fiber also has a simple way; I guess all other frameworks have it too; but their way of chaining will require a bit of boilerplate as well in the end, once the project scales; the only situation where the DX is faster is in the beginning or with smaller-scale projects;
  - so I'd say it's a tie in the end, since either net/http or frameworks would require a configurable mechanism to chain middleware in a scalable way; I'm leaning towards frameworks though, so + frameworks
  - the middleware chain boilerplate is in internal/middleware/middleware-chain.go; this method enables us to define an array/slice of middleware functions and create a chain; the handlers execute in the order they're defined in the array, so that works; a similar mechanism would be needed in frameworks as well
  - ok maybe it's a + for both net/http and frameworks
- subrouting - defining a router that handles a specific path prefix and then defining routes inside that router; for example, having a router for all "/users" routes and then defining "/users", "/users/{id}", "/users/items" inside that router; this is a must-have feature for any serious project
  - in net/http, we do this with a bit of extra boilerplate (which maybe can be abstracted away) by defining a new router with NewServeMux() and defining a .Handle("/prefix/", http.StripPrefix("/prefix", router)) where router is the router that handles the routes inside the prefix
  - defining a project structure to scale this would require some thought, but nothing colossal; it's a bit harder to think in a functional way when coming from JS/TS, or even Python, but it doesn't take long to get used to it
  - I'd say it's a half + for net/http since it works but it requires some organization; frameworks are a bit easier (at least Fiber), yet they will require some organization at higher scales as well; net/http enforces it right at the beginning, which is good practice let's say; so full + for frameworks
  - note: subrouting is also useful since we can define different middleware stacks to be applied to different routers; so maybe we have a /public/... series of endpoints which are public and don't require authentication, and then we have /admin/... which requires admin authentication; that's a use-case
- passing data through Context - this one I initially though was weird and I didn't hear about it outside the video, but in reality there is one common use-case where we see the need for context: authentication, mostly authorization
  - in tRPC and especially the t3 stack (Nextjs), we often have a context object we get to check user data; that context is essentially what the video author says here; that's an example of passing a context around to be able to identify the user and check if they have the necessary permissions
  - this situation can occur in any framework, any backend, of course; so it's not a net/http-specific thing
  - the proposed way we pass data here is by using the context package; every http request has an associated context, which we can extend to add the data we want
  - some underlying details I don't like here is that adding to the context (and the request by extension) creates new objects, so we're talking about some lost performance; the type-unsafe way of adding data to the context is also weird, since have the code to tell us what type the value is, and we also have the "value, ok" mechanism to tell us if there's something there or not
- (not a thing related to net/http, but Go) - auto-restart /watching for code changes; Nodejs has --watch and before this there was nodemon; there's plenty of options I've seen on a quick search; doesn't look like there's an industry standard; air looks nice

### Conclusion / thoughts overall

- based on my experience with Express, here's a list of features any reasonable project would need (as far as my experience takes me) and a checkmark if net/http has it or not + a blue square if there's a bit more boilerplate involved:
  - route handling ✅
  - method-based routing ✅
  - query parameters handling ✅
  - path parameters handling ✅
  - middleware & chaining ✅ 🟦
  - subrouting ✅ 🟦
  - cookie handling ✅
  - testing? - not sure what Go offers for this;
- I haven't included context cause I'd say it's not necessarily web-specific
- an important mention is that the JS ecosystem in general relies on 3rd party packages, while net/http is part of the standard library of Go; so we're as feature-rich as possible
- validation, session handling, database communication etc. in Nodejs are done through 3rd party packages unless you're insane; so I can't speak for Go being inferior because it doesn't have a validation package or anything; validation is just a middleware with specific functionality after all, so it's implementable; same with session and everything else
- so the overall vibe of Go is that everything is more barebones, we get the basic tools and it's our job to choose separate packages or build our own tools; in Js/Nodejs we rely on 3rd part packages for everything cause there is something available for everything and they're good; in Go, the situation is inconclusive (to me at least, for now)
- what I can say for sure is that DX is faster with JS and its frameworks; the tradeoff is performance, Go is way faster at CPU-intensive tasks; if we're talking about simple, IO-bound tasks, I'd say forget Go (unless you know it well cause then you waste time learning Express when you could be building)
- as for the starting question of net/http vs frameworks, I'd say it comes down to performance and DX; performance is supposed to be better on net/http since it's the barebones thing frameworks build upon, though I see Fiber prides itself with an up to 10x better performance and it has better DX (I'm biased due to Express); so I'd go with that unless I have specific situations where only net/http works
- if we're talking about feature completeness, net/http is there; everything else can be built upon it with relatively-low boilerplate overhead; so my answer is _yes_, net/http is enough for most if not all web server needs; the question now is project and person-specific: is net/http or a framework like Fiber better for my (team's) situation?

## Guess the number - TCP/UDP/QUIC?

-- Resources:

- [https://ops.tips/blog/udp-client-and-server-in-go/](https://ops.tips/blog/udp-client-and-server-in-go/) - this one feels in-depth and advanced
- [https://okanexe.medium.com/the-complete-guide-to-tcp-ip-connections-in-golang-1216dae27b5a](https://okanexe.medium.com/the-complete-guide-to-tcp-ip-connections-in-golang-1216dae27b5a)
- [https://quic-go.net/docs/](https://quic-go.net/docs/)

- Requirement: client-server architecture handling multiple clients at the same time (concurrency); when the server starts, a random number 1-10 is chosen; clients connect to the server and try to guess the number; the server responds with "too high", "too low", "correct!"; when a client guesses the number, the server changes it and prints an informative message to the log; when a client guesses correctly (receives "correct!"), it ends its execution
- HTTP is built on top of TCP/IP; with HTTP/3.0, it's now using another protocol called QUIC, built on top of UDP for its speed but borrowing the assurance of TCP
- TCP works by establishing a (secure if TLS) connection and using that to send data back; depends on version of HTTP being used, nowadays it's 2.0; not sure if I can force HTTP/1.1 with stdlib; this ensures the packet makes it back to the client
- with UDP there's no proper connection being established; there's no guarantee the packet makes it; packets are mostly lost when the network (or CPU in a local experiment) is busy
- !**INTERESTING FIND**: []byte of size 1024 (so 1024 byte slice) when converted to string via simple string(byteslice) will still keep its zero-valued part; if I try to do strconv.Atoi on that, I get an error; so make sure in the case of reading from connections into a buffer and then trying to convert that to only keep what's needed in the string part (so string(byteslice[:num])); sneaky bug right there, not obvious because we assume string(slice) only keeps what's needed due to it being the simplest way of converting []byte to string; alternative is fmt.Sprintf("%s",byteslice[:num]), but it still requires the limitation of bytes to be moved into the string; so either way string(...) is the easiest way of converting []byte into string
- rather tough to simulate & stress test simultaneous clients sending guesses; the code looks like it should handle things as intended; to test things out thoroughly, I should automate the sending process with random number guesses and automate the client starting with a bash script / goroutines; goroutines are easier

  - there's no deadlock with 100 clients and we know for sure due to the rwMutex that client responses are not stale
  - testing with 1000 clients introduces some dropped communication; doesn't seem to be due to deadlock, although Go does announce it as that; it seems to be just dropped packets due to overusage, since the server continues functioning
  - to test this, we can let the server expect a fixed number of clients; if at the end the server finishes, we know it's a package drop thing due to CPU usage
  - this testing leads to the actual answer - there is some deadlock; the server does not correctly close after accepting 1000 clients; is it due to all of them attempting near-simultaneous connection?
  - bug discovered, we weren't waiting for the goroutines to end on the server; introducng a WaitGroup and some mutex-covered counters for handled clients reveals clean handling at 10k clients with 1ms delay
  - yet having no delay, even if we use DialTimeout and allow 1s for the TCP connection to be established, the server misses a lot (9997 in one case) of clients; given the delay solves this issue, the root cause can be concluded as the server not being able to simultaneously listen to all clients;
  - (rookie mistake) printing the error reveals the issue; the connection is dropped by the server after accepting the client; so it's not able to hold thousands of connexions simultaneously; having the delay simply means we constantly clear out some requests to make way for others

- UDP: since UDP doesn't work by establishing connections, we essentially send one-off packets to the server; the response comes by using the sender's address to know where to send the packet to;

  - we'll see if many concurrent clients cause packet loss and we'll devise a way to retry sending packets to ensure eventual responses come
  - !**INTERESTING FIND**: using waitgroups - if we give the responsibility of marking a goroutine as waitable via wg.Add(n) to the goroutine, the main thread has a big chance of passing by the wg.Wait() call and so the goroutine(s) never get waited on; so it's a pattern to always add to the waitgroup in the main thread to ensure waitability
  - with a simple back-n-forth like with TCP, the program deadlocks (or has a chance of doing so at least) already at 10 simultaneous clients; sometimes, it's the server that continues to listen (so packets sent client->server were lost); sometimes, it's clients waiting for responses (so server->client packets were lost)
  - testing with delay between client launches of 1ms introduces even more issues; since UDP is not connection-based, the initial listen for requests in the main thread cannibalizes the listen on each client thread (goroutine); this is an indication UDP should not simulate connection-based communication by dedicating goroutines to each received message; maybe just handling the message itself; leaving the code as-is for future inspection purposes in commit **3695e03b21c0c547d70e8dd62b2872ea9d536a2d**, also called _UDP code+notes_
  - refactoring moving forward so that the goroutines are one-off handles of client packets; also changing server for loop to keep listening while we haven't met our client # expectations; this mechanism also invalidates the use of WaitGroups, since we don't have a fixed number of packets to read and handle;
  - having the server run as long as the number of clients (numClients) served with correct guesses doesn't match the expected number of clients also introduces the issue of being stuck in a read while the last client finishes their guess; so we introduce a deadline for reading as well, to refresh the server's tries so that we can gracefully exit
  - weirdly enough, this method (along with 1ms client delay) correctly processes 10k clients
  - removing the 1ms client delay introduces client deadlock, where the client(s) expect response(s) but they were lost along the way; so server->client packet loss
  - to introduce robustness, we can implement a retry mechanism for the client; this effectively solves our problem, since we mimick TCP's connection-based messaging with incorporated retries; the only issue is we need to make sure we don't cause an infinite loop; this lets us handle 10k clients with no delay (not guaranteed, but testing a few times did not introduce deadlocks)
  - there is still an issue here; if a client guesses correctly, but the response from the server gets lost, when the client resends its guess the server will have already incremented its numClients variable; this means that lost "Correct!" packages count multiple times; at some point a deadlock will occur;
  - testing with 100k clients introduces a new error on the client-side: i/o timeout; server slowed down to a halt, taking 7s at some point to process the packets

- what can we learn from this TCP/UDP situation? TCP ensures communication makes it to its destination, but sometimes connections drop so we need to reconnect; this happens in high-traffic situations; UDP does not ensure communication makes it, so we need to resend packets sometimes; overall, for one-directional communication UDP is likely the best candidate; this covers situations in which we send update, rather than requests; TCP is the better option when we expect responses, but we need to watch out for connection reset/timeout; there's a reason HTTP was built on top of TCP and not UDP
- these past few years, QUIC has been introduced and it is tightly coupled with HTTP 3.0; there is a library in Go supporting HTTP 3.0 and QUIC packets; let's explore that as well
  - using the quic-go package; it has docs to guide us a bit and we can also find information online; it's not official, looks like the standard library implementation of QUIC and HTTP 3.0 is still underway
  - based on the information in the docs and the [stackoverflow post here](https://stackoverflow.com/questions/77553334/why-is-this-toy-go-quic-server-accepting-connections-but-not-streams-when-the-q) it looks like we are able to establish our QUIC connection
  - !**IMPORTANT MENTION** the one who opens the stream is the one who must write first; in our case, the client establishes the connection and sends the guess, so it must open the bidirectional stream; QUIC allows for multiple streams over the same connection, so technically we could play guess-the-game over a single connection for multiple client goroutines (I guess it wouldn't work with 10k instances though)
  - as a small greeting test, the server opens the stream and sends a mesage to the client first; this is in the commit **2029a3af64e430651396470963409f076f4b72ba**, called "QUIC greeting"; **note - the stream.Write(...) function doesn't seem to block; if we don't put a delay to give the stream time to flush the message to the network, the server does not receive the message; 1ms is enough; lowest I got reliably is 100 microseconds**; I can't find enough information to provide a reliable solution, except for waiting
  - based on the documentation, the expected behaviour would be to get the message and then have the stream close; at the same time, modifying the code to only send from the server and receive on the client (without delaying on the server) does get the message across to the client; so the issue is only when sending from the client
