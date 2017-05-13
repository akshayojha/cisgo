# cisgo
A distributed Continuous Integration System written in Go for fun.

This is a basic distributed Continuous Integration System written in Golang for
Go projects hosted on Github. Once run it will watch the project repository on
regular interval for new commits. If a new commit is found it will run tests for
it and store the result in the file named as the commit hash.

# Explaination

The project consists of three main components:
- Scheduler : schedules the test for new commits
- Watcher : watches the given repository
- Tester : runs the test and reply back with the results
