# Make commands:
To build the program: make build
To clean the binaries: make clean

# Program arguments:
-p : Port number for TCP socket. UPD socket listens to the next port
-h : hostfile
-pause : Seconds to sleep upon starting

# Pass following flags to leader process to simulate test cases 2 and 4
-t4 : Executes logic for testcase 4
-t2 : Executes logic for testcase 2. Will not remove failed node


# Test Case 1:
Run `make start` on all the machines sequentially

# Test Case 2:
Run `make leader-test2` on the first leader machine.
Run `make start` on other machines.
Kill process (with CTRL + C) on each machine except for the leader

# Test Case 3:
Run `make start` on all the machines sequentially.
Kill process (with CTRL + C) on each machine except for the leader

# Test Case 4:
Run `make leader-test4` on the first leader machine.
Run `make start` on other machines.
Kill process (with CTRL + C) on any machine.
Leader will automatically exit once it detects the killed process.
