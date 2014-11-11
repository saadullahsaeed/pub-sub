# pub-sub 

_pub-sub_ is a simple Go library that allows you to implement a simple Publish-Subscribe mechanism. 

It uses go-routines and is of course asynchronous. 

### Go version 
This library assumes Go version 1.3.3+. Previous versions of Go have a different way of approaching GOPATH, etc, hence Makefile would have to be done differently. 

## The API

The API exposes 2 interfaces and 3 functions:
+ Sender -- which represents the Publisher of messages. Each Sender needs to be registered with a particular Receiver. 
+ Receiver -- which represents the Subscriber of messages. Receivers can register Senders, which means Publish invocations result in event execution in the Receiver. 


