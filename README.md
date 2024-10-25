# Exploring net/http package in Go to build web servers. Is it enough, or do we need frameworks?

And how easy is it to use? Does it require a lot of boilerplate?  
Following [this video](https://www.youtube.com/watch?v=H7tbjKFSg58) to understand new features; if net/http ends up being as feature rich as Express, I call no need for frameworks. I have another repo where I'm following an in-depth Express framework tutorial and go over all important features

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
  - - for net/http and + for frameworks due to these 2 issues of middleware chaining and obtaining status code...**is what I would say but I did not find an easy way of obtaining status code on Chi for example**; Chi has a simple way of chaining middleware; Fiber also has a simple way; I guess all other frameworks have it too; but their way of chaining will require a bit of boilerplate as well in the end, once the project scales; the only situation where the DX is faster is in the beginning or with smaller-scale projects;
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

### Conclusion / overall

- based on my experience with Express, here's a list of features any reasonable project would need (there may be more, I'm not an expert):
  - todo
