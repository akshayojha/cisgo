# cisgo
A distributed Continuous Integration System written in Go for fun.

This is a basic distributed Continuous Integration System written in Golang for
Go projects hosted on Github. Once run it will watch the project repository on
regular interval for new commits. If a new commit is found it will run tests for
it and store the result in the file named as the commit hash.

# Explanation

The project consists of three main components:
- Scheduler : schedules the test for new commits
- Watcher : watches the given repository
- Tester : runs the test and reply back with the results

The control flow is pretty simple too. The Watcher communicates with the Scheduler
when it finds out a new commit. It then passes the commit hash to the Scheduler,
which in turn looks for idle Tester and passes it the hash to run the tests.

Initially, the first component that starts is Scheduler which listens on a certain
port. The Watcher is fed the port number of the Scheduler, so that it knows whom
to communicate in case a new commit has been made. Similarly, the testers are fed
the address of the Scheduler to whom they register and report the results.

The project has an important utility component which contains common functions to
be used by all three components.

# Usage

To download use:

go get github.com/akshayojha/cisgo

- Navigate into your $GOPATH/src/github.com/akshayojha/cisgo
- Run make
- Navigate into you $GOPATH/bin/ and run the following:
