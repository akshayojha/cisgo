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
be used by all three components. Also, the watcher and each tester need a locally
stored copy of the github repository, which they use to observe and run tests from.

# Usage

To download use:

go get github.com/akshayojha/cisgo

- Navigate into your $GOPATH/src/github.com/akshayojha/cisgo
- Run make
- Navigate into you $GOPATH/bin/ and run in the following order:

Run the scheduler at desired ip and port

Usage of ./scheduler: <br/>
-sip string <br/>
    IP address of the scheduler server (default "localhost") <br/>
-sport string <br/>
    Port of the scheduler server (default "8080") <br/>

Run the tester and provide it the address of the Scheduler and also a local copy
of the git repository

Usage of ./tester: <br/>
-rpath string <br/>
  	Path to the repository folder to run tests from <br/>
-sip string <br/>
  	IP address of the scheduler server (default "localhost") <br/>
-sport string <br/>
  	Port of the scheduler server (default "8080")<br/>
-tip string<br/>
  	IP address of the tester (default "localhost")<br/>
-tport string<br/>
  	Port of the tester<br/>

Run the watcher finally and provide it the address of the Scheduler along with
another local copy of the repository

Usage of ./watcher:<br/>
  -rpath string<br/>
    	Path to the repository folder to observe<br/>
  -sip string<br/>
    	IP address of the scheduler server (default "localhost")<br/>
  -sport string<br/>
    	Port of the scheduler server (default "8080")<br/>

# TODO

- Add logic to implement post commit hooks of github so that we don't have to
watch repository all the time

- Improve performance and fault tolerance

- Make the code configurable by using a json formatted config file

# Known Bugs

- Scheduler is not fault tolerant. If it terminates every other component goes down
